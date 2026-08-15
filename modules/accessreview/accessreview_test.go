package accessreview

import "testing"

// A scope decides which rows a campaign copies out of RBAC, so a scope that
// names nothing is a campaign that silently asks about everything.
func TestScopeRules(t *testing.T) {
	cases := []struct {
		name  string
		in    Input
		ok    bool
		refIs string
	}{
		{"all needs no reference", Input{Name: "Q3", Scope: "all"}, true, ""},
		{"an empty scope means all", Input{Name: "Q3"}, true, ""},
		{"all discards a reference it was given", Input{Name: "Q3", Scope: "all", ScopeRef: "documents"}, true, ""},
		{"an app scope needs one", Input{Name: "Q3", Scope: "app"}, false, ""},
		{"an app scope with one", Input{Name: "Q3", Scope: "app", ScopeRef: "documents"}, true, "documents"},
		{"a role scope needs one", Input{Name: "Q3", Scope: "role"}, false, ""},
		{"an unknown scope", Input{Name: "Q3", Scope: "everything"}, false, ""},
		{"a campaign needs a name", Input{Scope: "all"}, false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.in
			err := in.validate()
			if tc.ok && err != nil {
				t.Fatalf("refused: %v", err)
			}
			if !tc.ok {
				if err == nil {
					t.Fatal("accepted")
				}
				return
			}
			if in.ScopeRef != tc.refIs {
				t.Fatalf("scope_ref is %q, want %q", in.ScopeRef, tc.refIs)
			}
		})
	}
}

func TestTheDueDateHasToBeADate(t *testing.T) {
	in := Input{Name: "Q3", DueDate: "next Friday"}
	if err := in.validate(); err == nil {
		t.Fatal("a due date that is not a date was accepted")
	}
	ok := Input{Name: "Q3", DueDate: "2026-09-30"}
	if err := ok.validate(); err != nil {
		t.Fatalf("a real date was refused: %v", err)
	}
}
