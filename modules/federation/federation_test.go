package federation

import "testing"

// What an issuer is allowed to be.
//
// This is the check worth a test: an OIDC issuer has discovery appended to it,
// so a value with a path or a trailing slash produces a URL that 404s at
// somebody's first sign-in — a long way from the screen where it was typed.
func TestIssuerRules(t *testing.T) {
	cases := []struct {
		name   string
		issuer string
		ok     bool
	}{
		{"an origin", "https://id.example.mn", true},
		{"a trailing slash is trimmed rather than refused", "https://id.example.mn/", true},
		{"a port is fine", "https://id.example.mn:8443", true},
		{"http is not", "http://id.example.mn", false},
		{"a path is not an issuer", "https://id.example.mn/realms/mn", false},
		{"nor is a query", "https://id.example.mn?tenant=1", false},
		{"nor is a bare host", "id.example.mn", false},
		{"nor is nothing", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := Input{DisplayName: "Ministry", ClientID: "abc", ClientSecret: "s", Issuer: tc.issuer}
			err := in.validate(true)
			if tc.ok && err != nil {
				t.Fatalf("%q was refused: %v", tc.issuer, err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("%q was accepted", tc.issuer)
			}
		})
	}
}

func TestTheDefaultsFillThemselvesIn(t *testing.T) {
	in := Input{DisplayName: "  Ministry  ", Issuer: "https://id.example.mn", ClientID: " abc "}
	if err := in.validate(false); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if in.DisplayName != "Ministry" || in.ClientID != "abc" {
		t.Fatalf("not trimmed: %q %q", in.DisplayName, in.ClientID)
	}
	if in.Scopes != "openid profile email" {
		t.Fatalf("scopes: %q", in.Scopes)
	}
	if in.AttributeMap == nil {
		t.Fatal("the attribute map should be an empty map, not nil: it is marshalled straight to JSONB")
	}
}

// An edit may omit the secret — that is what lets a console change a provider
// it can never read back — and a registration may not.
func TestTheSecretIsRequiredOnlyWhenRegistering(t *testing.T) {
	in := Input{DisplayName: "Ministry", Issuer: "https://id.example.mn", ClientID: "abc"}
	if err := in.validate(true); err == nil {
		t.Fatal("a registration without a secret was accepted")
	}
	if err := in.validate(false); err != nil {
		t.Fatalf("an edit without a secret was refused: %v", err)
	}
}
