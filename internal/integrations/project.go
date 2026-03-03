package integrations

import (
	"os"
	"path/filepath"
)

func DetectSupabaseProject(start string) (string, bool) {
	cur, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		supaDir := filepath.Join(cur, "supabase")
		if st, err := os.Stat(supaDir); err == nil && st.IsDir() {
			return cur, true
		}
		cfg := filepath.Join(cur, "config.toml")
		if _, err := os.Stat(cfg); err == nil {
			return cur, true
		}
		next := filepath.Dir(cur)
		if next == cur {
			return "", false
		}
		cur = next
	}
}
