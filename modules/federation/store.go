/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

package federation

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/jackc/pgx/v5"
)

// ErrNotFound is what the store answers for a row this tenant cannot see —
// which is the same answer as for a row that does not exist, deliberately. A
// distinguishable "exists but not yours" tells a caller what other
// organisations have registered.
var ErrNotFound = errors.New("no such provider")

// Provider is a federation as the API describes it.
//
// There is no field for the client secret. It is written through Create and
// Update and read only by the code that uses it; a struct with a place to put
// it is a struct somebody will eventually marshal.
type Provider struct {
	ID           string            `json:"id"`
	DisplayName  string            `json:"display_name"`
	Issuer       string            `json:"issuer"`
	ClientID     string            `json:"client_id"`
	Scopes       string            `json:"scopes"`
	AttributeMap map[string]string `json:"attribute_map"`
	Enabled      bool              `json:"enabled"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
	// HasSecret says a credential is stored without saying anything about it.
	// The question a console actually asks is "is this one configured", and
	// answering it does not require handing back the answer's contents.
	HasSecret bool `json:"has_secret"`
}

// Link is one person's identity at one provider.
//
// ProviderName and Email are filled on the organisation-wide listing and left
// empty on a provider's own. The screen that asks "who has arrived through
// federation at all" needs both to be readable; the one already looking at a
// provider knows its name, and a join for a column nobody reads is a join.
type Link struct {
	ID           string `json:"id"`
	ProviderID   string `json:"provider_id"`
	ProviderName string `json:"provider_name,omitempty"`
	Subject      string `json:"subject"`
	UserID       string `json:"user_id"`
	Email        string `json:"email,omitempty"`
	LinkedAt     string `json:"linked_at"`
}

type Store struct{ db nexus.DB }

func NewStore(db nexus.DB) *Store { return &Store{db: db} }

const providerColumns = `id, display_name, issuer, client_id, scopes, attribute_map, enabled,
	to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'), to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
	octet_length(client_secret_encrypted) > 0`

func (s *Store) List(ctx context.Context, tenantID string) ([]Provider, error) {
	rows, err := s.db.Query(ctx, `SELECT `+providerColumns+`
		FROM sso_federation_providers WHERE tenant_id = $1 ORDER BY display_name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	providers := []Provider{}
	for rows.Next() {
		provider, err := scanProvider(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, rows.Err()
}

func (s *Store) Get(ctx context.Context, tenantID, id string) (Provider, error) {
	row := s.db.QueryRow(ctx, `SELECT `+providerColumns+`
		FROM sso_federation_providers WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	provider, err := scanProvider(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Provider{}, ErrNotFound
	}
	return provider, err
}

// scanRow is what a pgx.Row and a pgx.Rows have in common, so one scan serves
// the single-row read and the loop.
type scanRow interface{ Scan(dest ...any) error }

func scanProvider(row scanRow) (Provider, error) {
	var provider Provider
	var attributes []byte
	if err := row.Scan(&provider.ID, &provider.DisplayName, &provider.Issuer, &provider.ClientID,
		&provider.Scopes, &attributes, &provider.Enabled,
		&provider.CreatedAt, &provider.UpdatedAt, &provider.HasSecret); err != nil {
		return Provider{}, err
	}
	provider.AttributeMap = map[string]string{}
	if len(attributes) > 0 {
		// A map that will not unmarshal is a stored value somebody wrote by
		// hand; it should not take the whole list down with it.
		_ = json.Unmarshal(attributes, &provider.AttributeMap)
	}
	return provider, nil
}

// Create registers a provider and returns it as the API will describe it.
func (s *Store) Create(ctx context.Context, tenantID, actorID string, in Input, sealed []byte) (Provider, error) {
	attributes, err := json.Marshal(in.AttributeMap)
	if err != nil {
		return Provider{}, err
	}
	var id string
	err = s.db.QueryRow(ctx, `
		INSERT INTO sso_federation_providers
			(tenant_id, display_name, issuer, client_id, client_secret_encrypted, scopes, attribute_map, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		tenantID, in.DisplayName, in.Issuer, in.ClientID, sealed, in.Scopes, attributes, nullUUID(actorID),
	).Scan(&id)
	if err != nil {
		return Provider{}, err
	}
	return s.Get(ctx, tenantID, id)
}

// Update replaces the editable fields. A nil sealed secret leaves the stored
// one alone, which is what lets a console edit a provider it can never read.
func (s *Store) Update(ctx context.Context, tenantID, id string, in Input, sealed []byte) (Provider, error) {
	attributes, err := json.Marshal(in.AttributeMap)
	if err != nil {
		return Provider{}, err
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE sso_federation_providers SET
			display_name = $3, issuer = $4, client_id = $5, scopes = $6, attribute_map = $7,
			client_secret_encrypted = COALESCE($8, client_secret_encrypted),
			updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`,
		tenantID, id, in.DisplayName, in.Issuer, in.ClientID, in.Scopes, attributes, sealed)
	if err != nil {
		return Provider{}, err
	}
	if tag.RowsAffected() == 0 {
		return Provider{}, ErrNotFound
	}
	return s.Get(ctx, tenantID, id)
}

func (s *Store) SetEnabled(ctx context.Context, tenantID, id string, enabled bool) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE sso_federation_providers SET enabled = $3, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`, tenantID, id, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, tenantID, id string) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM sso_federation_providers WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Links(ctx context.Context, tenantID, providerID string) ([]Link, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, provider_id, subject, user_id, to_char(linked_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM sso_federation_links
		WHERE tenant_id = $1 AND provider_id = $2
		ORDER BY linked_at DESC`, tenantID, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []Link{}
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.ProviderID, &link.Subject, &link.UserID, &link.LinkedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

// AllLinks answers who has arrived through federation at all, newest first.
//
// The per-provider listing above answers a question somebody already inside a
// provider is asking. This one answers the question an administrator opens the
// module with — which of our people sign in from outside, and through whom —
// and it is not the same query with a filter removed: it has to name the
// provider and the person, which the other does not.
func (s *Store) AllLinks(ctx context.Context, tenantID string) ([]Link, error) {
	rows, err := s.db.Query(ctx, `
		SELECT l.id, l.provider_id, p.display_name, l.subject, l.user_id, u.email,
			to_char(l.linked_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM sso_federation_links l
		JOIN sso_federation_providers p ON p.id = l.provider_id
		JOIN users u ON u.id = l.user_id
		WHERE l.tenant_id = $1
		ORDER BY l.linked_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []Link{}
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.ProviderID, &link.ProviderName, &link.Subject,
			&link.UserID, &link.Email, &link.LinkedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

// nullUUID turns an empty actor into SQL NULL. A zero UUID would be a user id
// that looks real and belongs to nobody.
func nullUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
