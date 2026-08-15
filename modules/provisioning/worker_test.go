package provisioning

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// A bearer token is sent on every request to a target, so what an operator may
// point one at is a security decision rather than a formatting one.
func TestBaseURLRules(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		ok   bool
	}{
		{"https://scim.example.mn/v2", "https://scim.example.mn/v2", true},
		{"https://scim.example.mn/v2/", "https://scim.example.mn/v2", true},
		{"  https://scim.example.mn  ", "https://scim.example.mn", true},
		{"http://scim.example.mn/v2", "", false},
		{"https://scim.example.mn/v2?tenant=1", "", false},
		{"scim.example.mn", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, err := validBaseURL(tc.raw)
		if tc.ok && err != nil {
			t.Errorf("%q was refused: %v", tc.raw, err)
			continue
		}
		if !tc.ok {
			if err == nil {
				t.Errorf("%q was accepted", tc.raw)
			}
			continue
		}
		if got != tc.want {
			t.Errorf("%q became %q, want %q", tc.raw, got, tc.want)
		}
	}
}

// The excerpt of a remote system's answer goes into the database, and the cut
// lands mid-rune the first time that answer is not in English.
func TestTruncateLeavesValidUTF8(t *testing.T) {
	long := strings.Repeat("хэрэглэгч", 200)
	got := truncate(long)
	if !utf8.ValidString(got) {
		t.Fatal("the excerpt is not valid UTF-8")
	}
	if len(got) > 520 {
		t.Fatalf("the excerpt is %d bytes", len(got))
	}
	if short := truncate("ok"); short != "ok" {
		t.Fatalf("a short answer was altered: %q", short)
	}
}
