// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package applications нь gateway service id ↔ OAuth scope хөрвүүлэгч.
//
// ӨМНӨ НЬ энэ багц applications / application_services хүснэгтүүдийг ч удирддаг
// байсан. Client-ууд бүрэн эхээрээ oauth_clients руу шилжсэн тул тэр хоёр
// хүснэгт болон тэднийг уншдаг код устсан (migration 6) — апп хаана
// бүртгэгддэг нь НЭГ л газар байх ёстой.
package applications

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	repointerface "template/internal/datasources/repositories/interface"
)

type applicationRepository struct {
	pool *pgxpool.Pool
}

func NewApplicationRepository(pool *pgxpool.Pool) repointerface.ApplicationRepository {
	return &applicationRepository{pool: pool}
}

// ServiceScopes нь өгсөн gateway service id-уудын OAuth scope нэрсийг буцаана.
func (r *applicationRepository) ServiceScopes(ctx context.Context, serviceIDs []string) ([]string, error) {
	if len(serviceIDs) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT scope FROM gateway_services WHERE id = ANY($1::uuid[]) AND scope <> ''`, serviceIDs)
	if err != nil {
		return nil, fmt.Errorf("query service scopes: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ServiceIDsForScopes нь өгсөн OAuth scope нэрсэд харгалзах gateway service
// id-уудыг буцаана (ServiceScopes-ийн урвуу). Client-ийн scope-оос апп-ын
// зөвшөөрсөн service-үүдийг сэргээхэд ашиглана.
func (r *applicationRepository) ServiceIDsForScopes(ctx context.Context, scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id::text FROM gateway_services WHERE scope = ANY($1) AND scope <> ''`, scopes)
	if err != nil {
		return nil, fmt.Errorf("query service ids for scopes: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
