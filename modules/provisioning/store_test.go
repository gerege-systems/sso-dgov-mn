package provisioning

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The tests that need rows.
//
// What they are about is the queue: claiming a job, finishing one, and the
// backoff that decides when a failure is tried again and when it is given up
// on. None of that can be checked without a database — a backoff computed in Go
// and written as an interval string is only correct if PostgreSQL agrees.
func testStore(t *testing.T) (*Store, *pgxpool.Pool) {
	t.Helper()
	dsn := os.Getenv("SSO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("SSO_TEST_DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return NewStore(pool), pool
}

// org builds an organisation with one member and one enabled target.
func org(t *testing.T, store *Store, pool *pgxpool.Pool) (string, string) {
	t.Helper()
	ctx := context.Background()
	var tenantID, userID string

	if err := pool.QueryRow(ctx, `
		INSERT INTO tenants (slug, name) VALUES ('t-' || gen_random_uuid(), 'Test') RETURNING id`).
		Scan(&tenantID); err != nil {
		t.Fatalf("tenant: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM tenants WHERE id = $1`, tenantID) })

	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name)
		VALUES (gen_random_uuid() || '@example.mn', 'x', 'Test Person') RETURNING id`).
		Scan(&userID); err != nil {
		t.Fatalf("user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID) })

	if _, err := pool.Exec(ctx,
		`INSERT INTO memberships (tenant_id, user_id) VALUES ($1, $2)`, tenantID, userID); err != nil {
		t.Fatalf("membership: %v", err)
	}

	target, err := store.Create(ctx, tenantID, userID,
		Input{Name: "Payroll", BaseURL: "https://scim.example.mn/v2"}, []byte("sealed"))
	if err != nil {
		t.Fatalf("target: %v", err)
	}
	return tenantID, target.ID
}

func TestAResyncQueuesEveryMemberAndTheWorkerCanClaimThem(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, targetID := org(t, store, pool)

	queued, err := store.EnqueueAll(ctx, tenantID, targetID)
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if queued != 1 {
		t.Fatalf("queued %d, expected the organisation's one member", queued)
	}

	tx, jobs, err := store.claim(ctx, 10)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	found := false
	for _, j := range jobs {
		if j.targetID == targetID {
			found = true
			if j.baseURL != "https://scim.example.mn/v2" || string(j.token) != "sealed" {
				t.Errorf("the claim did not carry the target's credentials: %+v", j)
			}
			if j.op != "update" || len(j.payload) == 0 {
				t.Errorf("the claim did not carry the work: %+v", j)
			}
		}
	}
	if !found {
		t.Fatal("the queued job was not claimed")
	}
}

// A disabled target's work waits rather than being sent. Turning a target off
// is how an operator stops it, and a queue that ignored the flag would keep
// pushing for as long as there was a backlog.
func TestADisabledTargetIsNotClaimed(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, targetID := org(t, store, pool)
	if _, err := store.EnqueueAll(ctx, tenantID, targetID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx,
		`UPDATE sso_scim_targets SET enabled = FALSE WHERE id = $1`, targetID); err != nil {
		t.Fatal(err)
	}

	tx, jobs, err := store.claim(ctx, 50)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, j := range jobs {
		if j.targetID == targetID {
			t.Fatal("a disabled target's work was claimed")
		}
	}
}

func TestFinishingRemovesTheJobAndLogsIt(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, targetID := org(t, store, pool)
	if _, err := store.EnqueueAll(ctx, tenantID, targetID); err != nil {
		t.Fatal(err)
	}

	tx, jobs, err := store.claim(ctx, 50)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	for _, j := range jobs {
		if j.targetID != targetID {
			continue
		}
		if err := store.finish(ctx, tx, j, 201, "created"); err != nil {
			t.Fatalf("finish: %v", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	var pending int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM sso_scim_queue WHERE target_id = $1`, targetID).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	if pending != 0 {
		t.Fatalf("%d jobs are still queued after finishing", pending)
	}

	runs, err := store.Runs(ctx, tenantID, 10)
	if err != nil {
		t.Fatalf("runs: %v", err)
	}
	if len(runs) != 1 || runs[0].StatusCode != 201 {
		t.Fatalf("the outcome was not logged: %+v", runs)
	}
}

// A failure waits, and waits longer each time, and eventually stops waiting.
func TestRetryBacksOffAndThenGivesUp(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, targetID := org(t, store, pool)
	if _, err := store.EnqueueAll(ctx, tenantID, targetID); err != nil {
		t.Fatal(err)
	}

	var previous time.Duration
	for attempt := 0; attempt < maxAttempts; attempt++ {
		tx, jobs, err := store.claim(ctx, 50)
		if err != nil {
			t.Fatalf("claim %d: %v", attempt, err)
		}
		var mine *job
		for i := range jobs {
			if jobs[i].targetID == targetID {
				mine = &jobs[i]
			}
		}
		if mine == nil {
			t.Fatalf("the job disappeared on attempt %d", attempt)
		}
		if mine.attempts != attempt {
			t.Fatalf("attempt counter is %d on round %d", mine.attempts, attempt)
		}
		if err := store.retry(ctx, tx, *mine, 503, "the target is unavailable"); err != nil {
			t.Fatalf("retry: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit: %v", err)
		}

		var wait *time.Duration
		var seconds *float64
		if err := pool.QueryRow(ctx, `
			SELECT EXTRACT(EPOCH FROM next_attempt_at - NOW())
			FROM sso_scim_queue WHERE target_id = $1`, targetID).Scan(&seconds); err != nil {
			// The last retry deletes the row, which is what gave up looks like.
			if attempt != maxAttempts-1 {
				t.Fatalf("the job vanished on attempt %d: %v", attempt, err)
			}
			break
		}
		next := time.Duration(*seconds * float64(time.Second))
		wait = &next
		if *wait <= previous {
			t.Fatalf("attempt %d waits %s, no longer than the previous %s", attempt, *wait, previous)
		}
		previous = *wait

		// Make it due again so the next round can claim it.
		if _, err := pool.Exec(ctx,
			`UPDATE sso_scim_queue SET next_attempt_at = NOW() WHERE target_id = $1`, targetID); err != nil {
			t.Fatal(err)
		}
	}

	var left int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM sso_scim_queue WHERE target_id = $1`, targetID).Scan(&left); err != nil {
		t.Fatal(err)
	}
	if left != 0 {
		t.Fatalf("the job is still queued after %d attempts", maxAttempts)
	}

	runs, _ := store.Runs(ctx, tenantID, 20)
	if len(runs) != maxAttempts {
		t.Fatalf("expected %d logged attempts, got %d", maxAttempts, len(runs))
	}
	if runs[0].Response == "" {
		t.Error("the last log line does not say what happened")
	}
}
