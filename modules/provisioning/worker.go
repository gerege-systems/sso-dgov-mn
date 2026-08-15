/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

package provisioning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// maxAttempts is where the backoff stops. Six attempts spread over about an
	// hour: long enough to ride out a restart at the other end, short enough
	// that a target which is genuinely broken shows up in the log today.
	maxAttempts = 6
	// batch is how many jobs one tick claims. Small, because the whole batch is
	// held in one transaction and a long transaction is a long lock.
	batch = 20
	// tick is how often the queue is looked at. Provisioning is not
	// interactive: nobody is watching a spinner while a user appears in another
	// system.
	tick = 30 * time.Second
)

// worker drains the queue.
//
// It runs inside this process rather than as a separate deployment, which is
// the same decision the platform makes for its own sweeps: a second thing to
// deploy is a second thing to forget to deploy, and this one does nothing at
// all on an installation with no targets.
type worker struct {
	store  *Store
	reveal func([]byte) (string, error)
	client *http.Client
}

func (w *worker) run(ctx context.Context) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.drain(ctx); err != nil {
				slog.Error("provisioning: the queue could not be drained", "error", err)
			}
		}
	}
}

// drain claims a batch, sends each job, and commits the outcomes together.
//
// One transaction for the batch means a crash mid-batch replays the whole
// batch. SCIM is not transactional and a replayed PUT is a second PUT of
// identical content, which is why every operation below is written to be safe
// to repeat: create is a lookup then a POST or a PUT, never a blind POST.
func (w *worker) drain(ctx context.Context) error {
	tx, jobs, err := w.store.claim(ctx, batch)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if len(jobs) == 0 {
		return tx.Commit(ctx)
	}

	for _, j := range jobs {
		status, body, sendErr := w.send(ctx, j)
		if sendErr != nil {
			body = sendErr.Error()
		}
		// 2xx is done. Everything else waits, including a 4xx: a target that
		// refuses a payload today because a required attribute is missing is a
		// target somebody will fix, and dropping the job would mean the fix
		// changes nothing until the next full resync.
		if sendErr == nil && status >= 200 && status < 300 {
			err = w.store.finish(ctx, tx, j, status, body)
		} else {
			err = w.store.retry(ctx, tx, j, status, body)
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// send performs one job against one target.
//
// The remote id is looked up by externalId rather than stored here. Storing it
// would mean this database holds a second opinion about what exists over there,
// and the two are wrong in different ways after somebody deletes a user in the
// target's own console.
func (w *worker) send(ctx context.Context, j job) (int, string, error) {
	token, err := w.reveal(j.token)
	if err != nil {
		return 0, "", fmt.Errorf("the target's token could not be read: %w", err)
	}

	remoteID, status, body, err := w.lookup(ctx, j, token)
	if err != nil {
		return status, body, err
	}

	var payload map[string]any
	if len(j.payload) > 0 {
		_ = json.Unmarshal(j.payload, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}
	payload["externalId"] = j.userID
	if j.op == "deactivate" {
		payload["active"] = false
	}

	switch {
	case remoteID == "" && j.op == "deactivate":
		// Nothing to deactivate. The desired state is already the state.
		return http.StatusOK, "not present at the target", nil
	case remoteID == "":
		return w.do(ctx, http.MethodPost, j.baseURL+"/Users", token, payload)
	default:
		return w.do(ctx, http.MethodPut, j.baseURL+"/Users/"+url.PathEscape(remoteID), token, payload)
	}
}

// lookup asks the target whether it already knows this person.
func (w *worker) lookup(ctx context.Context, j job, token string) (string, int, string, error) {
	query := url.Values{"filter": {`externalId eq "` + j.userID + `"`}}
	status, body, err := w.request(ctx, http.MethodGet, j.baseURL+"/Users?"+query.Encode(), token, nil)
	if err != nil {
		return "", status, body, err
	}
	if status == http.StatusNotFound {
		return "", status, body, nil
	}
	if status < 200 || status >= 300 {
		return "", status, body, fmt.Errorf("the target refused the lookup: %d", status)
	}

	var answer struct {
		Resources []struct {
			ID string `json:"id"`
		} `json:"Resources"`
	}
	if err := json.Unmarshal([]byte(body), &answer); err != nil {
		return "", status, body, fmt.Errorf("the target's lookup answer did not parse: %w", err)
	}
	if len(answer.Resources) == 0 {
		return "", status, body, nil
	}
	return answer.Resources[0].ID, status, body, nil
}

func (w *worker) do(ctx context.Context, method, endpoint, token string, payload map[string]any) (int, string, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return 0, "", err
	}
	return w.request(ctx, method, endpoint, token, encoded)
}

func (w *worker) request(ctx context.Context, method, endpoint, token string, body []byte) (int, string, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/scim+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/scim+json")
	}

	res, err := w.client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = res.Body.Close() }()

	// Bounded, because the body is a stranger's and is about to be written to
	// this database.
	answer, err := io.ReadAll(io.LimitReader(res.Body, 1<<16))
	if err != nil {
		return res.StatusCode, "", err
	}
	return res.StatusCode, string(answer), nil
}

// validBaseURL is what an operator is allowed to point a target at.
//
// https only, and no path fragment that would make the endpoint this code
// builds land somewhere unintended. A bearer token travels on every one of
// these requests; http would put it on the wire in the clear on the first sync
// and nobody would be watching.
func validBaseURL(raw string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("base_url must be a URL")
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("base_url must be https: a bearer token is sent with every request")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("base_url must have no query or fragment")
	}
	return trimmed, nil
}
