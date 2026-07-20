//go:build integration

// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Бүрэн authorize → consent → token → userinfo → refresh урсгалыг ЖИНХЭНЭ
// Postgres дээр, ЖИНХЭНЭ RLS-ийн доор (non-superuser app_user) ажиллуулна.
//
// ЯАГААД ЭНЭ ТЕСТ БАЙХ ЁСТОЙ: unit тестүүд нь санах-ойн хуурамч store ашигладаг
// тул RLS-ийг огт хөнддөггүй. Ингэснээр протоколын логик бүхэлдээ ногоон байтал
// production-д (болон compose-д) хүсэлт бүр RLS-д хаагдаж, provider огт
// ажиллахгүй байх боломжтой байсан — яг тийм алдаа гарч байсан.
package oidc_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"template/internal/business/domain"
	oidcuc "template/internal/business/usecases/oidc"
	usersuc "template/internal/business/usecases/users"
	oauthpg "template/internal/datasources/repositories/postgres/oauth"
	"template/internal/test/testenv"
	"template/pkg/secrethash"
)

const (
	testClientID = "ring-dgov-mn"
	testSecret   = "integration-test-client-secret"
	testRedirect = "https://ring.dgov.mn/sso/callback"
	testIssuer   = "https://sso.dgov.mn"
)

// stubUsers нь id_token/userinfo-д хэрэгтэй иргэний бүртгэлийг өгнө.
type stubUsers struct{ id string }

func (s *stubUsers) GetByID(context.Context, usersuc.GetByIDRequest) (usersuc.GetByIDResponse, error) {
	return usersuc.GetByIDResponse{User: domain.User{
		ID: s.id, FirstName: "Бат", LastName: "Дорж", Email: "bat@example.mn",
	}}, nil
}

// setup нь RLS хүчинтэй (non-superuser) pool дээр service-ийг угсарна.
func setup(t *testing.T) (*oidcuc.Service, string) {
	t.Helper()
	admin := testenv.StartPostgres(t)
	app := testenv.AppUserPool(t, admin)

	// Production-ийн initdb нь migrate-ийн үүсгэсэн бүх хүснэгтэд DML эрхийг
	// ALTER DEFAULT PRIVILEGES-ээр олгодог; харнес нь зөвхөн users-т олгодог тул
	// oauth_* хүснэгтүүдэд эрхийг гараар нэмнэ (RLS нь эрхийн ДЭЭР ажилладаг).
	for _, tbl := range []string{
		"oauth_clients", "oauth_signing_keys", "oauth_auth_codes",
		"oauth_access_tokens", "oauth_refresh_tokens", "oauth_challenges", "oauth_consents",
	} {
		_, err := admin.Exec(context.Background(),
			`GRANT SELECT, INSERT, UPDATE, DELETE ON `+tbl+` TO app_user`)
		require.NoError(t, err, "grant on %s", tbl)
	}

	subject := seedUser(t, admin)
	seedClient(t, app)

	flow := oauthpg.NewFlowRepository(app)
	clients := oauthpg.NewClientRepository(app)
	keys, err := oidcuc.NewKeyManager(oauthpg.NewKeyRepository(app), "integration-test-encryption-key")
	require.NoError(t, err)
	require.NoError(t, keys.EnsureKey(context.Background()))

	svc := oidcuc.NewService(clients, flow, testIssuer).
		WithTokenIssuing(keys, &stubUsers{id: subject})
	return svc, subject
}

func seedUser(t *testing.T, admin *pgxpool.Pool) string {
	t.Helper()
	// FK-д зориулж жинхэнэ хэрэглэгч мөр (superuser-ээр, RLS тойрч) — org_rls_test-
	// ийн адил хэв маяг.
	const id = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	_, err := admin.Exec(context.Background(),
		`INSERT INTO users(id, username, active, role_id, created_at)
		 VALUES ($1, 'oidc_test', true, 4, now())`, id)
	require.NoError(t, err)
	return id
}

func seedClient(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	hash, err := secrethash.Hash(testSecret)
	require.NoError(t, err)
	_, err = oauthpg.NewClientRepository(pool).Create(context.Background(), domain.OAuthClient{
		ClientID:                testClientID,
		ClientName:              "ring.dgov.mn",
		SecretHash:              hash,
		TokenEndpointAuthMethod: domain.AuthMethodBasic,
		AppType:                 "web",
		GrantTypes:              []string{domain.GrantAuthorizationCode, domain.GrantRefreshToken},
		ResponseTypes:           []string{"code"},
		Scopes:                  []string{"openid", "profile", "email", "offline_access"},
		RedirectURIs:            []string{testRedirect},
		PostLogoutRedirectURIs:  []string{"https://ring.dgov.mn/"},
		Enabled:                 true,
	})
	require.NoError(t, err)
}

// Бүтэн урсгал: authorize → login accept → consent accept → code exchange →
// userinfo → refresh. Алхам бүр RLS-ийн доор жинхэнэ DB-д бичиж/уншина.
func TestFullAuthorizationCodeFlowUnderRLS(t *testing.T) {
	svc, subject := setup(t)
	ctx := context.Background()

	// PKCE: verifier → S256 challenge.
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challengeParam := oidcuc.S256Challenge(verifier)

	loginChallenge, client, err := svc.Authorize(ctx, oidcuc.AuthorizeRequest{
		ClientID:            testClientID,
		RedirectURI:         testRedirect,
		ResponseType:        "code",
		Scope:               "openid profile email offline_access",
		State:               "state-abc",
		Nonce:               "nonce-xyz",
		CodeChallenge:       challengeParam,
		CodeChallengeMethod: "S256",
	})
	require.NoError(t, err, "authorize must work under RLS")
	require.NotEmpty(t, loginChallenge)
	require.Equal(t, testClientID, client.ClientID)

	consentChallenge, skip, err := svc.AcceptLogin(ctx, loginChallenge, subject)
	require.NoError(t, err)
	require.False(t, skip, "first time through there is no remembered consent")

	redirect, err := svc.AcceptConsent(ctx, consentChallenge, subject, nil)
	require.NoError(t, err)

	u, err := url.Parse(redirect)
	require.NoError(t, err)
	code := u.Query().Get("code")
	require.NotEmpty(t, code, "redirect must carry a code: %s", redirect)
	require.Equal(t, "state-abc", u.Query().Get("state"))

	tok, err := svc.Token(ctx, oidcuc.TokenRequest{
		GrantType: domain.GrantAuthorizationCode, Code: code,
		RedirectURI: testRedirect, CodeVerifier: verifier,
		ClientID: testClientID, ClientSecret: testSecret, SecretFromBasic: true,
	})
	require.NoError(t, err, "code exchange must work under RLS")
	require.NotEmpty(t, tok.AccessToken)
	require.NotEmpty(t, tok.RefreshToken, "offline_access was granted")
	require.NotEmpty(t, tok.IDToken, "openid was granted")

	// id_token нь энэ issuer, энэ client, энэ subject-д зориулагдсан байх ёстой.
	require.Equal(t, 3, len(strings.Split(tok.IDToken, ".")), "id_token must be a JWS")

	info := svc.Introspect(ctx, testClientID, tok.AccessToken)
	require.True(t, info.Active)
	require.Equal(t, subject, info.Subject)

	claims, err := svc.Userinfo(ctx, tok.AccessToken)
	require.NoError(t, err)
	require.Equal(t, subject, claims["sub"])
	require.Equal(t, "Дорж Бат", claims["name"])

	// Refresh нь эргэлт хийж, хуучныг хүчингүй болгоно.
	refreshed, err := svc.Token(ctx, oidcuc.TokenRequest{
		GrantType: domain.GrantRefreshToken, RefreshToken: tok.RefreshToken,
		ClientID: testClientID, ClientSecret: testSecret, SecretFromBasic: true,
	})
	require.NoError(t, err, "refresh must work under RLS")
	require.NotEqual(t, tok.RefreshToken, refreshed.RefreshToken, "refresh token must rotate")

	// Хуучин refresh-ийг дахин ашиглах → бүлэг цуцлагдана.
	_, err = svc.Token(ctx, oidcuc.TokenRequest{
		GrantType: domain.GrantRefreshToken, RefreshToken: tok.RefreshToken,
		ClientID: testClientID, ClientSecret: testSecret, SecretFromBasic: true,
	})
	require.Error(t, err, "a consumed refresh token must be refused")

	require.False(t, svc.Introspect(ctx, testClientID, refreshed.AccessToken).Active,
		"detecting refresh reuse must revoke the whole family, including the newest access token")
}

// Хоёр дахь удаагийн нэвтрэлт: санагдсан зөвшөөрөл хүссэн бүх scope-ыг хамарвал
// consent UI алгасагдана.
func TestRememberedConsentSkipsUnderRLS(t *testing.T) {
	svc, subject := setup(t)
	ctx := context.Background()

	run := func() (string, bool) {
		loginChallenge, _, err := svc.Authorize(ctx, oidcuc.AuthorizeRequest{
			ClientID: testClientID, RedirectURI: testRedirect, ResponseType: "code",
			Scope: "openid profile",
		})
		require.NoError(t, err)
		consentChallenge, skip, err := svc.AcceptLogin(ctx, loginChallenge, subject)
		require.NoError(t, err)
		return consentChallenge, skip
	}

	consent, skip := run()
	require.False(t, skip)
	_, err := svc.AcceptConsent(ctx, consent, subject, nil)
	require.NoError(t, err)

	_, skip = run()
	require.True(t, skip, "the remembered grant covers the request, so consent should be skipped")
}
