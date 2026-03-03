package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileVaultRoundTrip(t *testing.T) {
	t.Setenv("SUPASHIFT_PASSPHRASE", "test-passphrase-123")
	dir := t.TempDir()
	v := NewFileVault(filepath.Join(dir, "vault.age"))

	if err := v.SetToken("p1", "tok_abc"); err != nil {
		t.Fatalf("set token: %v", err)
	}
	got, err := v.GetToken("p1")
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if got != "tok_abc" {
		t.Fatalf("unexpected token: %s", got)
	}
	if err := v.DeleteToken("p1"); err != nil {
		t.Fatalf("delete token: %v", err)
	}
	if _, err := v.GetToken("p1"); err == nil {
		t.Fatalf("expected missing token error")
	}

	if _, err := os.Stat(filepath.Join(dir, "vault.age")); err != nil {
		t.Fatalf("vault file missing: %v", err)
	}
}
