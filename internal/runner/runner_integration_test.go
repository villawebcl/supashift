package runner

import "testing"

func TestRunInjectsToken(t *testing.T) {
	err := Run("prof-a", "abc123", []string{"sh", "-c", `[ "$SUPABASE_ACCESS_TOKEN" = "abc123" ] && [ "$SUPASHIFT_ACTIVE_PROFILE" = "prof-a" ]`})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
