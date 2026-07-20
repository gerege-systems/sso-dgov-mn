package verifymig

import (
	"os/exec"
	"strings"
	"testing"

	"template/pkg/secrethash"
)

// Зөөгдсөн hash-ууд нь RP-үүдийн ОДООГИЙН secret-ээр шалгагдах ёстой.
func TestMigratedHashesVerify(t *testing.T) {
	want := map[string]string{
		"template-dgov-mn": "tmpl-secret-value-123",
		"ring-dgov-mn":     "ring-secret-value-456",
	}
	out, err := exec.Command("docker", "exec", "-i", "mig-db", "psql", "-U", "postgres", "-d", "appdb",
		"-At", "-F", "\t", "-c", "select client_id, secret_hash from oauth_clients order by client_id").Output()
	if err != nil {
		t.Fatalf("read migrated clients: %v", err)
	}
	seen := 0
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		id, hash := parts[0], parts[1]
		secret, ok := want[id]
		if !ok {
			continue // public client — secret байхгүй
		}
		seen++
		got, err := secrethash.Verify(hash, secret)
		if err != nil {
			t.Fatalf("%s: Verify errored on a migrated hash: %v", id, err)
		}
		if !got {
			t.Fatalf("%s: the RP's existing secret NO LONGER validates after migration — cutover would break it", id)
		}
		if ok, _ := secrethash.Verify(hash, secret+"x"); ok {
			t.Fatalf("%s: a wrong secret validated", id)
		}
	}
	if seen != len(want) {
		t.Fatalf("checked %d confidential clients, expected %d", seen, len(want))
	}
}
