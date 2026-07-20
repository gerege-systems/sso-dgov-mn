// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package oauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"template/internal/apperror"
	"template/internal/business/domain"
	"template/internal/datasources/rls"
)

// flowRepository нь authorize урсгалын түр төлөвийг (challenge, санагдсан
// consent, authorization code) хадгална.
//
// Эдгээр хүснэгтүүд RLS-тэй бөгөөд протоколын endpoint-ууд нэвтрэхээс ӨМНӨ
// ажилладаг тул query бүр withRLS транзакцаар (ихэвчлэн RoleService) явна.
type flowRepository struct {
	pool *pgxpool.Pool
}

func NewFlowRepository(pool *pgxpool.Pool) *flowRepository {
	return &flowRepository{pool: pool}
}

// withRLS нь context дахь Identity-г SET LOCAL болгож тавина (users repository-
// тэй ижил загвар). Identity байхгүй бол GUC хоосон → бодлого бүх мөрийг хаана.
func (r *flowRepository) withRLS(ctx context.Context, fn func(tx pgx.Tx) error) error {
	id, _ := rls.FromContext(ctx)
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // commit хийсний дараах rollback нь ErrTxClosed — хүлээгдсэн
	if _, err := tx.Exec(ctx,
		`SELECT set_config('app.user_id',$1,true), set_config('app.user_role',$2,true)`,
		id.UserID, string(id.Role),
	); err != nil {
		return fmt.Errorf("set rls session context: %w", err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ── Challenge ────────────────────────────────────────────────────────────────

const challengeColumns = ` challenge, kind, client_id, subject, requested_scopes, granted_scopes,
	redirect_uri, state, nonce, response_type, code_challenge, code_challenge_method,
	prompt, post_logout_redirect_uri, skip, decided_at, expires_at, created_at`

func scanChallenge(row pgx.Row) (domain.OAuthChallenge, error) {
	var c domain.OAuthChallenge
	err := row.Scan(&c.Challenge, &c.Kind, &c.ClientID, &c.Subject, &c.RequestedScopes, &c.GrantedScopes,
		&c.RedirectURI, &c.State, &c.Nonce, &c.ResponseType, &c.CodeChallenge, &c.CodeChallengeMethod,
		&c.Prompt, &c.PostLogoutRedirectURI, &c.Skip, &c.DecidedAt, &c.ExpiresAt, &c.CreatedAt)
	return c, err
}

// CreateChallenge нь шинэ login/consent/logout challenge бичнэ.
func (r *flowRepository) CreateChallenge(ctx context.Context, c domain.OAuthChallenge) error {
	return r.withRLS(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO oauth_challenges (
				challenge, kind, client_id, subject, requested_scopes, granted_scopes,
				redirect_uri, state, nonce, response_type, code_challenge, code_challenge_method,
				prompt, post_logout_redirect_uri, skip, expires_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
			c.Challenge, c.Kind, nullableText(c.ClientID), nullableUUID(c.Subject),
			strList(c.RequestedScopes), strList(c.GrantedScopes),
			c.RedirectURI, c.State, c.Nonce, c.ResponseType, c.CodeChallenge, c.CodeChallengeMethod,
			c.Prompt, c.PostLogoutRedirectURI, c.Skip, c.ExpiresAt)
		if err != nil {
			return fmt.Errorf("insert oauth challenge: %w", err)
		}
		return nil
	})
}

// Challenge нь ХҮЧИНТЭЙ (хугацаа дуусаагүй, хараахан шийдэгдээгүй) challenge-ыг
// буцаана. Дуусал/шийдэгдсэнийг NotFound-той ижилхэн үзнэ — дахин ашиглахаас
// сэргийлнэ.
func (r *flowRepository) Challenge(ctx context.Context, kind, challenge string) (domain.OAuthChallenge, error) {
	var out domain.OAuthChallenge
	err := r.withRLS(ctx, func(tx pgx.Tx) error {
		c, scanErr := scanChallenge(tx.QueryRow(ctx,
			`SELECT`+challengeColumns+` FROM oauth_challenges
			 WHERE challenge = $1 AND kind = $2 AND decided_at IS NULL AND expires_at > now()`,
			challenge, kind))
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return apperror.NotFound("challenge not found or already used")
		}
		if scanErr != nil {
			return fmt.Errorf("get oauth challenge: %w", scanErr)
		}
		out = c
		return nil
	})
	return out, err
}

// DecideChallenge нь challenge-ыг шийдэгдсэн гэж тэмдэглэнэ (нэг удаагийн).
// Хэрэв өөр хүсэлт аль хэдийн шийдсэн бол NotFound — давхар зарцуулалт болохгүй.
func (r *flowRepository) DecideChallenge(ctx context.Context, challenge, subject string, granted []string) error {
	return r.withRLS(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE oauth_challenges
			   SET decided_at = now(), subject = COALESCE($2, subject), granted_scopes = $3
			 WHERE challenge = $1 AND decided_at IS NULL AND expires_at > now()`,
			challenge, nullableUUID(subject), strList(granted))
		if err != nil {
			return fmt.Errorf("decide oauth challenge: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return apperror.NotFound("challenge not found or already used")
		}
		return nil
	})
}

// ── Санагдсан consent ────────────────────────────────────────────────────────

// Consent нь тухайн иргэн тухайн апп-д өмнө нь олгосон, хүчинтэй scope-уудыг
// буцаана. Байхгүй бол хоосон (алдаа биш) — consent UI харуулна.
func (r *flowRepository) Consent(ctx context.Context, subject, clientID string) ([]string, error) {
	var scopes []string
	err := r.withRLS(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx,
			`SELECT scopes FROM oauth_consents
			  WHERE subject = $1 AND client_id = $2 AND expires_at > now()`,
			subject, clientID)
		if scanErr := row.Scan(&scopes); scanErr != nil {
			if errors.Is(scanErr, pgx.ErrNoRows) {
				scopes = nil
				return nil
			}
			return fmt.Errorf("get oauth consent: %w", scanErr)
		}
		return nil
	})
	return scopes, err
}

// SaveConsent нь олгосон scope-уудыг санана (дараагийн удаа UI алгасна).
func (r *flowRepository) SaveConsent(ctx context.Context, subject, clientID string, scopes []string, ttl time.Duration) error {
	return r.withRLS(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO oauth_consents (subject, client_id, scopes, expires_at)
			VALUES ($1, $2, $3, now() + $4::interval)
			ON CONFLICT (subject, client_id) DO UPDATE
			   SET scopes = EXCLUDED.scopes, expires_at = EXCLUDED.expires_at, updated_at = now()`,
			subject, clientID, strList(scopes), ttl.String())
		if err != nil {
			return fmt.Errorf("save oauth consent: %w", err)
		}
		return nil
	})
}

// RevokeConsent нь тухайн апп-д олгосон зөвшөөрлийг устгана.
func (r *flowRepository) RevokeConsent(ctx context.Context, subject, clientID string) error {
	return r.withRLS(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`DELETE FROM oauth_consents WHERE subject = $1 AND client_id = $2`, subject, clientID)
		if err != nil {
			return fmt.Errorf("revoke oauth consent: %w", err)
		}
		return nil
	})
}

// ── Authorization code ───────────────────────────────────────────────────────

// CreateCode нь authorization code-ыг (hash хэлбэрээр) бичнэ.
func (r *flowRepository) CreateCode(ctx context.Context, c domain.OAuthAuthCode) error {
	return r.withRLS(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO oauth_auth_codes (
				code_hash, client_id, subject, scopes, redirect_uri, nonce,
				code_challenge, code_challenge_method, auth_time, expires_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			c.CodeHash, c.ClientID, c.Subject, strList(c.Scopes), c.RedirectURI, c.Nonce,
			c.CodeChallenge, c.CodeChallengeMethod, c.AuthTime, c.ExpiresAt)
		if err != nil {
			return fmt.Errorf("insert authorization code: %w", err)
		}
		return nil
	})
}

// ConsumeCode нь code-ыг АТОМААР нэг удаа зарцуулна: хугацаа дуусаагүй бөгөөд
// хараахан ашиглагдаагүй бол used_at тавьж мөрийг буцаана.
//
// Хоёр дахь удаа ирвэл (`used` = true) дуудагч энэ нь дахин ашиглалт гэдгийг
// мэдэж, тухайн session-ий бүх token-ыг цуцлах ёстой (RFC 6749 §4.1.2).
func (r *flowRepository) ConsumeCode(ctx context.Context, codeHash []byte) (code domain.OAuthAuthCode, alreadyUsed bool, err error) {
	err = r.withRLS(ctx, func(tx pgx.Tx) error {
		// Эхлээд мөрийг түгжиж уншина — өрсөлдөөнт солилцоо давхар амжилтгүй болно.
		var c domain.OAuthAuthCode
		var usedAt *time.Time
		scanErr := tx.QueryRow(ctx, `
			SELECT code_hash, client_id, subject, scopes, redirect_uri, nonce,
			       code_challenge, code_challenge_method, auth_time, expires_at, used_at
			  FROM oauth_auth_codes
			 WHERE code_hash = $1
			 FOR UPDATE`, codeHash).Scan(
			&c.CodeHash, &c.ClientID, &c.Subject, &c.Scopes, &c.RedirectURI, &c.Nonce,
			&c.CodeChallenge, &c.CodeChallengeMethod, &c.AuthTime, &c.ExpiresAt, &usedAt)
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return apperror.NotFound("authorization code not found")
		}
		if scanErr != nil {
			return fmt.Errorf("get authorization code: %w", scanErr)
		}

		code = c
		if usedAt != nil {
			alreadyUsed = true
			return nil // дуудагч бүлгийг цуцална
		}
		if time.Now().After(c.ExpiresAt) {
			return apperror.BadRequest("authorization code expired")
		}

		if _, execErr := tx.Exec(ctx,
			`UPDATE oauth_auth_codes SET used_at = now() WHERE code_hash = $1`, codeHash); execErr != nil {
			return fmt.Errorf("consume authorization code: %w", execErr)
		}
		return nil
	})
	return code, alreadyUsed, err
}

// DeleteExpired нь хугацаа дууссан түр мөрүүдийг цэвэрлэнэ (тогтмол ажил).
// Ашиглагдсан code-ыг ХЭСЭГ ХУГАЦААНД үлдээнэ — дахин ашиглалтыг илрүүлэхэд
// хэрэгтэй; зөвхөн эрс хуучирсныг нь хаяна.
func (r *flowRepository) DeleteExpired(ctx context.Context) error {
	return r.withRLS(ctx, func(tx pgx.Tx) error {
		for _, q := range []string{
			`DELETE FROM oauth_challenges WHERE expires_at < now() - interval '1 day'`,
			`DELETE FROM oauth_auth_codes WHERE expires_at < now() - interval '1 day'`,
			`DELETE FROM oauth_access_tokens WHERE expires_at < now() - interval '7 days'`,
			`DELETE FROM oauth_refresh_tokens WHERE expires_at < now() - interval '7 days'`,
			`DELETE FROM oauth_consents WHERE expires_at < now()`,
		} {
			if _, err := tx.Exec(ctx, q); err != nil {
				return fmt.Errorf("cleanup oauth state: %w", err)
			}
		}
		return nil
	})
}

func nullableText(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableUUID(s string) any {
	if s == "" {
		return nil
	}
	return s
}
