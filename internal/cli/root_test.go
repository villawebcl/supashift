package cli

import (
	"strings"
	"testing"
)

func TestAutoSetSnippetEscapesProfile(t *testing.T) {
	profile := `client-'$HOME-$(id)-\` + "`uname`"
	got := autoSetSnippet(profile)

	wantPart := `supashift use -- 'client-'\''$HOME-$(id)-\` + "`uname`" + `'`
	if !strings.Contains(got, wantPart) {
		t.Fatalf("auto snippet does not escape profile safely:\n%s", got)
	}
	if !strings.HasPrefix(got, `eval "$(supashift use -- `) {
		t.Fatalf("unexpected prefix: %q", got)
	}
}
