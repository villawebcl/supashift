package runner

import "testing"

func TestUseSnippetShellEscapesValues(t *testing.T) {
	profile := "client-'$HOME-`uname`-$(id)"
	token := "tok-'$HOME-`uname`-$(id)"

	snippet := UseSnippet(profile, token)
	checkCmd := snippet +
		`test "$SUPABASE_ACCESS_TOKEN" = ` + shellSingleQuote(token) +
		` && test "$SUPASHIFT_ACTIVE_PROFILE" = ` + shellSingleQuote(profile)

	err := Run("check", "safe", []string{
		"sh",
		"-c",
		checkCmd,
	})
	if err != nil {
		t.Fatalf("snippet did not preserve literal values: %v", err)
	}
}
