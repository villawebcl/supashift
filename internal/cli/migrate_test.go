package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadLegacySupabaseToken(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "access-token")
	if err := os.WriteFile(p, []byte(" tok_file \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SUPABASE_ACCESS_TOKEN", "tok_env")

	tok, src, err := readLegacySupabaseToken("auto", p, "SUPABASE_ACCESS_TOKEN")
	if err != nil {
		t.Fatalf("auto err: %v", err)
	}
	if tok != "tok_file" || src != "file" {
		t.Fatalf("auto unexpected: tok=%q src=%q", tok, src)
	}

	tok, src, err = readLegacySupabaseToken("env", p, "SUPABASE_ACCESS_TOKEN")
	if err != nil {
		t.Fatalf("env err: %v", err)
	}
	if tok != "tok_env" || src != "env" {
		t.Fatalf("env unexpected: tok=%q src=%q", tok, src)
	}
}
