// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"errors"
	"fmt"

	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"template/internal/apperror"
	"template/internal/business/domain"
	"template/internal/datasources/records"
	"template/pkg/logger"
)

// UpsertSuperAdmin нь superadmin onboarding-ийн ТӨГСГӨЛД (Google + eID + email
// OTP + TOTP бүгд баталгаажсаны дараа) super admin хэрэглэгчийг үүсгэх/ахиулна.
//
// Түлхүүр нь UpsertFromEID-ийн адил civil_id (idx_users_civil_id_active partial
// unique index, migration 13): тухайн иргэн аль хэдийн eID-ээр нэвтэрч байсан
// бол түүний мөрийг ахиулж (role_id, email + email_verified, mfa_enabled,
// totp_secret, Google профайл), эс бөгөөс шинэ идэвхтэй super admin мөр
// оруулна. Нэг round-trip (INSERT … ON CONFLICT … RETURNING).
//
// АНХААР: totp_secret нь usecase давхаргад AES-GCM-ээр аль хэдийн шифрлэгдсэн
// ирнэ — энэ давхаргад ил текст secret ХЭЗЭЭ Ч бичигдэхгүй. Латин нэрийг
// UpsertFromEID-ийн адил COALESCE-оор нэг л удаа авна (хэрэглэгчийн гар
// засварыг дарж бичихгүй). Нууц үг NULL — super admin нь Google + TOTP-оор л
// нэвтэрнэ.
func (r *postgreUserRepository) UpsertSuperAdmin(ctx context.Context, inDom *domain.User) (domain.User, error) {
	const (
		repositoryName = "users"
		funcName       = "UpsertSuperAdmin"
		queryName      = "upsertSuperAdmin"
		fileName       = "users_superadmin.go"
	)
	rec := records.FromUsersV1Domain(inDom)

	var stored records.Users
	err := r.withRLS(ctx, func(tx pgx.Tx) error {
		rows, qErr := tx.Query(ctx, `
			INSERT INTO users(
				id, username, first_name, last_name, first_name_en, last_name_en,
				email, password, active, role_id, national_id, civil_id, kyc_level,
				google_sub, google_email, google_email_verified, google_name, google_picture, google_linked_at,
				email_verified, mfa_enabled, totp_secret, created_at)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5,
				$6, NULL, true, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, now(),
				true, true, $16, $17)
			ON CONFLICT (lower(civil_id)) WHERE civil_id IS NOT NULL
			DO UPDATE SET
				first_name            = EXCLUDED.first_name,
				last_name             = EXCLUDED.last_name,
				-- Латин нэрийг НЭГ УДАА (анхны insert-д) л авна; хэрэглэгчийн
				-- гараар засварласан утгыг дарж бичихгүй (UpsertFromEID-ийн адил).
				first_name_en         = COALESCE(users.first_name_en, EXCLUDED.first_name_en),
				last_name_en          = COALESCE(users.last_name_en, EXCLUDED.last_name_en),
				national_id           = EXCLUDED.national_id,
				kyc_level             = EXCLUDED.kyc_level,
				email                 = EXCLUDED.email,
				email_verified        = true,
				mfa_enabled           = true,
				totp_secret           = EXCLUDED.totp_secret,
				google_sub            = EXCLUDED.google_sub,
				google_email          = EXCLUDED.google_email,
				google_email_verified = EXCLUDED.google_email_verified,
				google_name           = EXCLUDED.google_name,
				google_picture        = EXCLUDED.google_picture,
				google_linked_at      = COALESCE(users.google_linked_at, now()),
				role_id               = EXCLUDED.role_id,
				active                = true,
				updated_at            = now()
			RETURNING `+records.UserColumns+`
		`,
			rec.Username, rec.FirstName, rec.LastName, rec.FirstNameEn, rec.LastNameEn,
			rec.Email, rec.RoleId, rec.NationalID, rec.CivilID, rec.KYCLevel,
			rec.GoogleSub, rec.GoogleEmail, rec.GoogleEmailVerified, rec.GoogleName, rec.GooglePicture,
			rec.TOTPSecret, rec.CreatedAt)
		if qErr != nil {
			return qErr
		}
		var scanErr error
		stored, scanErr = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[records.Users])
		return scanErr
	})
	if err == nil {
		if stored.Id == "" {
			e := fmt.Errorf("upsert succeeded but RETURNING produced no row")
			logger.ErrorWithContext(ctx, "superadmin upsert returned no row", logger.Fields{
				"repository": repositoryName, "method": funcName, "query": queryName,
				"file": fileName, "error": e.Error(), "table": "users",
			})
			return domain.User{}, apperror.InternalCause(e)
		}
		return stored.ToV1Domain(), nil
	}

	// Урьсан и-мэйл / Google account өөр бүртгэлд аль хэдийн эзэмшигдсэн бол
	// (idx_users_email_active / idx_users_google_sub_active) цэвэр 409 болгоно.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		logger.ErrorWithContext(ctx, "superadmin upsert conflict", logger.Fields{
			"repository": repositoryName, "method": funcName, "query": queryName,
			"file": fileName, "constraint": pgErr.ConstraintName, "table": "users",
		})
		return domain.User{}, apperror.Conflict("this email or Google account is already linked to another user")
	}
	logger.ErrorWithContext(ctx, "Failed to upsert super admin", logger.Fields{
		"repository": repositoryName, "method": funcName, "query": queryName,
		"file": fileName, "error": err.Error(), "table": "users",
	})
	return domain.User{}, apperror.InternalCause(fmt.Errorf("upsert super admin: %w", err))
}
