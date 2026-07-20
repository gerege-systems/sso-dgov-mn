// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Discovery болон JWKS нь ГАДААД гэрээ — RP-ийн сангууд болон iOS апп эдгээрийг
// татаж, endpoint болон алгоритмаа тохируулдаг. Тиймээс энэ тест нь HTTP-ийн
// бодит биетийг (JSON) шалгана, дотоод бүтцийг биш.
package oidc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"template/internal/apperror"
	"template/internal/business/domain"
	oidcuc "template/internal/business/usecases/oidc"
	oidchandler "template/internal/http/handlers/v1/oidc"
)

const testIssuer = "https://sso.dgov.mn"

type memKeys struct{ keys []domain.SigningKey }

func (m *memKeys) Active(context.Context) (domain.SigningKey, error) {
	for _, k := range m.keys {
		if k.Active {
			return k, nil
		}
	}
	return domain.SigningKey{}, apperror.NotFound("no active signing key")
}
func (m *memKeys) All(context.Context) ([]domain.SigningKey, error) { return m.keys, nil }
func (m *memKeys) Insert(_ context.Context, k domain.SigningKey) error {
	m.keys = append([]domain.SigningKey{k}, m.keys...)
	return nil
}
func (m *memKeys) RetireActive(context.Context) error {
	for i := range m.keys {
		m.keys[i].Active = false
	}
	return nil
}

func newHandler(t *testing.T) oidchandler.Handler {
	t.Helper()
	km, err := oidcuc.NewKeyManager(&memKeys{}, "handler-test-encryption-key")
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	if err := km.EnsureKey(context.Background()); err != nil {
		t.Fatalf("EnsureKey: %v", err)
	}
	return oidchandler.NewHandler(km, testIssuer)
}

func TestDiscoveryDocument(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler(t).Discovery(rec, httptest.NewRequest(http.MethodGet, oidcuc.PathDiscovery, http.NoBody))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json;charset=UTF-8" {
		t.Fatalf("Content-Type = %q", ct)
	}

	var d map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatalf("discovery is not valid JSON: %v", err)
	}

	// Endpoint-ууд issuer дээр суурилсан бүтэн URL байх ёстой — RP-үүд эдгээрийг
	// шууд ашигладаг тул харьцангуй зам ажиллахгүй.
	wantURLs := map[string]string{
		"issuer":                 testIssuer,
		"authorization_endpoint": testIssuer + "/oauth2/auth",
		"token_endpoint":         testIssuer + "/oauth2/token",
		"userinfo_endpoint":      testIssuer + "/userinfo",
		"jwks_uri":               testIssuer + "/.well-known/jwks.json",
		"revocation_endpoint":    testIssuer + "/oauth2/revoke",
		"end_session_endpoint":   testIssuer + "/oauth2/sessions/logout",
	}
	for k, want := range wantURLs {
		if got, _ := d[k].(string); got != want {
			t.Fatalf("%s = %q, want %q", k, got, want)
		}
	}

	// PKCE: ЗӨВХӨН S256. "plain" зарлавал халдагч challenge-ыг дур мэдэн
	// сонгож PKCE-ийн хамгаалалтыг үгүй болгоно (RFC 9700 §2.1.1).
	methods := toStrings(d["code_challenge_methods_supported"])
	if len(methods) != 1 || methods[0] != "S256" {
		t.Fatalf("code_challenge_methods_supported = %v; must advertise S256 only", methods)
	}

	// Хэрэгжүүлээгүй урсгалыг зарлаж болохгүй.
	for _, g := range toStrings(d["grant_types_supported"]) {
		if g == "implicit" || g == "password" {
			t.Fatalf("grant_types_supported advertises %q which is not implemented", g)
		}
	}
	if rt := toStrings(d["response_types_supported"]); len(rt) != 1 || rt[0] != "code" {
		t.Fatalf("response_types_supported = %v; only the code flow is implemented", rt)
	}

	// Дотоод gateway service-ийн нэрс НИЙТЭД гарах ёсгүй.
	for _, s := range toStrings(d["scopes_supported"]) {
		if len(s) > 4 && s[:4] == "svc:" {
			t.Fatalf("scopes_supported leaks the internal gateway scope %q", s)
		}
	}

	if algs := toStrings(d["id_token_signing_alg_values_supported"]); len(algs) != 1 || algs[0] != "RS256" {
		t.Fatalf("id_token_signing_alg_values_supported = %v, want [RS256]", algs)
	}
	if subs := toStrings(d["subject_types_supported"]); len(subs) != 1 || subs[0] != "public" {
		t.Fatalf("subject_types_supported = %v, want [public]", subs)
	}
}

func TestJWKSEndpoint(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler(t).JWKS(rec, httptest.NewRequest(http.MethodGet, oidcuc.PathJWKS, http.NoBody))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var set struct {
		Keys []map[string]any `json:"keys"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &set); err != nil {
		t.Fatalf("jwks is not valid JSON: %v", err)
	}
	if len(set.Keys) != 1 {
		t.Fatalf("expected one published key, got %d", len(set.Keys))
	}

	k := set.Keys[0]
	for _, field := range []string{"kty", "use", "alg", "kid", "n", "e"} {
		if v, ok := k[field].(string); !ok || v == "" {
			t.Fatalf("jwk is missing %q: %+v", field, k)
		}
	}

	// Хувийн түлхүүрийн бүрэлдэхүүн ХЭЗЭЭ Ч JWKS-д гарч болохгүй.
	for _, private := range []string{"d", "p", "q", "dp", "dq", "qi"} {
		if _, leaked := k[private]; leaked {
			t.Fatalf("JWKS leaked the private component %q", private)
		}
	}
}

func toStrings(v any) []string {
	raw, _ := v.([]any)
	out := make([]string, 0, len(raw))
	for _, x := range raw {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
