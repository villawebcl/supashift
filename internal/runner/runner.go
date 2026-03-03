package runner

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func sanitizedEnv(profile, token string) []string {
	base := make([]string, 0, len(os.Environ())+2)
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "SUPABASE_ACCESS_TOKEN=") {
			continue
		}
		base = append(base, kv)
	}
	base = append(base, "SUPABASE_ACCESS_TOKEN="+token)
	base = append(base, "SUPASHIFT_ACTIVE_PROFILE="+profile)
	return base
}

func Run(profile, token string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("comando vacío")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = sanitizedEnv(profile, token)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func defaultShell() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	if sh := os.Getenv("SHELL"); sh != "" {
		return sh
	}
	return "zsh"
}

func Shell(profile, token string) error {
	shell := defaultShell()
	fmt.Fprintf(os.Stderr, "[supashift] Perfil activo: %s\n", profile)
	cmd := exec.Command(shell)
	cmd.Env = append(sanitizedEnv(profile, token), "SUPASHIFT_BANNER=1")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Tmux(profile, token string) error {
	return tmux(profile, token, true)
}

func TmuxDetached(profile, token string) error {
	return tmux(profile, token, false)
}

func tmux(profile, token string, attach bool) error {
	safe := regexp.MustCompile(`[^a-zA-Z0-9_-]+`).ReplaceAllString(profile, "-")
	session := "supa-" + safe
	newCmd := exec.Command("tmux", "new-session", "-Ad", "-s", session,
		"-e", "SUPABASE_ACCESS_TOKEN="+token,
		"-e", "SUPASHIFT_ACTIVE_PROFILE="+profile)
	newCmd.Stdin = os.Stdin
	newCmd.Stdout = os.Stdout
	newCmd.Stderr = os.Stderr
	if err := newCmd.Run(); err != nil {
		return err
	}
	if !attach {
		return nil
	}
	att := exec.Command("tmux", "attach-session", "-t", session)
	att.Stdin = os.Stdin
	att.Stdout = os.Stdout
	att.Stderr = os.Stderr
	return att.Run()
}

func UseSnippet(profile, token string) string {
	return fmt.Sprintf("export SUPABASE_ACCESS_TOKEN=%q\nexport SUPASHIFT_ACTIVE_PROFILE=%q\necho 'supashift: perfil activo %s'\n", token, profile, profile)
}

func UnuseSnippet() string {
	return "unset SUPABASE_ACCESS_TOKEN\nunset SUPASHIFT_ACTIVE_PROFILE\necho 'supashift: entorno limpiado'\n"
}
