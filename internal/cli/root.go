package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/villawebcl/supashift/internal/config"
	"github.com/villawebcl/supashift/internal/integrations"
	"github.com/villawebcl/supashift/internal/model"
	"github.com/villawebcl/supashift/internal/runner"
	"github.com/villawebcl/supashift/internal/tui"
	"github.com/villawebcl/supashift/internal/vault"
)

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "supashift",
		Short: "Gestiona perfiles Supabase CLI con sesiones aisladas",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return showFirstRunBanner()
		},
	}

	root.AddCommand(newInitCmd())
	root.AddCommand(newProfileCmd())
	root.AddCommand(newRunCmd())
	root.AddCommand(newShellCmd())
	root.AddCommand(newTmuxCmd())
	root.AddCommand(newPickCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newExportCmd())
	root.AddCommand(newImportCmd())
	root.AddCommand(newMigrateCmd())
	root.AddCommand(newUseCmd())
	root.AddCommand(newUnuseCmd())
	root.AddCommand(newAutoCmd())
	root.AddCommand(newProjectCmd())
	root.AddCommand(newRevealCmd())
	root.AddCommand(newCompletionsCmd())
	root.AddCommand(newSnippetCmd())

	return root
}

func showFirstRunBanner() error {
	if os.Getenv("SUPASHIFT_NO_BANNER") == "1" {
		return nil
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil
	}
	dir, err := config.EnsureDirs()
	if err != nil {
		return nil
	}
	marker := filepath.Join(dir, ".welcome_seen")
	if _, err := os.Stat(marker); err == nil {
		return nil
	}
	banner := `
  ____                        _     _  __ _
 / ___| _   _ _ __   __ _ ___| |__ (_)/ _| |_
 \___ \| | | | '_ \ / _` + "`" + ` / __| '_ \| | |_| __|
  ___) | |_| | |_) | (_| \__ \ | | | |  _| |_
 |____/ \__,_| .__/ \__,_|___/_| |_|_|_|  \__|
             |_|

supashift: perfiles Supabase aislados por sesión.
quickstart:
  supashift init
  supashift profile add <perfil> --account "<correo>"
  supashift run <perfil> -- supabase projects list
`
	_, _ = io.WriteString(os.Stderr, banner)
	_ = os.WriteFile(marker, []byte("seen\n"), 0o600)
	return nil
}

func loadCfgVault() (*model.Config, *vault.Manager, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	dir, err := config.ConfigDir()
	if err != nil {
		return nil, nil, err
	}
	mgr, err := vault.NewManager(dir, cfg.VaultBackend)
	if err != nil {
		return nil, nil, err
	}
	return cfg, mgr, nil
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Inicializa configuración local",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := config.InitConfig()
			if err != nil {
				return err
			}
			fmt.Printf("Config inicializada en %s (vault_backend=%s)\n", path, cfg.VaultBackend)
			return nil
		},
	}
}

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "profile", Short: "Gestiona perfiles"}
	cmd.AddCommand(newProfileAddCmd(), newProfileEditCmd(), newProfileRmCmd(), newProfileLsCmd())
	return cmd
}

func parseCSV(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func shellSingleQuote(v string) string {
	return "'" + strings.ReplaceAll(v, "'", `'\''`) + "'"
}

func autoSetSnippet(profile string) string {
	return fmt.Sprintf("eval \"$(supashift use -- %s)\"\n", shellSingleQuote(profile))
}

func readTokenInteractive() (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New("usa --token o ejecuta en terminal interactiva")
	}
	fmt.Fprint(os.Stderr, "Token Supabase (input oculto): ")
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(string(b))
	if v == "" {
		return "", errors.New("token vacío")
	}
	return v, nil
}

func newProfileAddCmd() *cobra.Command {
	var account, notes, tags, aliases, token string
	var favorite bool
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Agrega perfil",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg, mgr, err := loadCfgVault()
			if err != nil {
				return err
			}
			if _, ok := cfg.Profiles[name]; ok {
				return fmt.Errorf("perfil ya existe: %s", name)
			}
			if token == "" {
				token, err = readTokenInteractive()
				if err != nil {
					return err
				}
			}
			p := model.Profile{
				Name:         name,
				AccountLabel: account,
				Notes:        notes,
				Tags:         parseCSV(tags),
				Aliases:      parseCSV(aliases),
				Favorite:     favorite,
			}
			config.UpsertProfile(cfg, p)
			if err := config.Save(cfg); err != nil {
				return err
			}
			if err := mgr.Vault().SetToken(name, token); err != nil {
				return err
			}
			fmt.Printf("Perfil %s agregado (backend=%s)\n", name, mgr.Vault().Backend())
			return nil
		},
	}
	cmd.Flags().StringVar(&account, "account", "", "Etiqueta de cuenta")
	cmd.Flags().StringVar(&notes, "notes", "", "Notas")
	cmd.Flags().StringVar(&tags, "tags", "", "Tags separadas por coma")
	cmd.Flags().StringVar(&aliases, "aliases", "", "Aliases separadas por coma")
	cmd.Flags().StringVar(&token, "token", "", "Token PAT (evita pasarlo directo; usa prompt)")
	cmd.Flags().BoolVar(&favorite, "favorite", false, "Marcar como favorito")
	_ = cmd.MarkFlagRequired("account")
	return cmd
}

func newProfileEditCmd() *cobra.Command {
	var account, notes, tags, aliases, token string
	var favorite bool
	var setFavorite bool
	cmd := &cobra.Command{
		Use:   "edit <profile>",
		Short: "Edita perfil",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, mgr, err := loadCfgVault()
			if err != nil {
				return err
			}
			name, p, err := config.ResolveProfile(cfg, args[0])
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("account") {
				p.AccountLabel = account
			}
			if cmd.Flags().Changed("notes") {
				p.Notes = notes
			}
			if cmd.Flags().Changed("tags") {
				p.Tags = parseCSV(tags)
			}
			if cmd.Flags().Changed("aliases") {
				p.Aliases = parseCSV(aliases)
			}
			if setFavorite {
				p.Favorite = favorite
			}
			config.UpsertProfile(cfg, p)
			if err := config.Save(cfg); err != nil {
				return err
			}
			if cmd.Flags().Changed("token") {
				if token == "" {
					tk, err := readTokenInteractive()
					if err != nil {
						return err
					}
					token = tk
				}
				if err := mgr.Vault().SetToken(name, token); err != nil {
					return err
				}
			}
			fmt.Printf("Perfil %s actualizado\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&account, "account", "", "Etiqueta")
	cmd.Flags().StringVar(&notes, "notes", "", "Notas")
	cmd.Flags().StringVar(&tags, "tags", "", "Tags")
	cmd.Flags().StringVar(&aliases, "aliases", "", "Aliases")
	cmd.Flags().StringVar(&token, "token", "", "Nuevo token")
	cmd.Flags().BoolVar(&favorite, "favorite", false, "Favorito true/false")
	cmd.Flags().BoolVar(&setFavorite, "set-favorite", false, "Aplicar valor de --favorite")
	return cmd
}

func newProfileRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <profile>",
		Short: "Elimina perfil",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, mgr, err := loadCfgVault()
			if err != nil {
				return err
			}
			name, _, err := config.ResolveProfile(cfg, args[0])
			if err != nil {
				return err
			}
			delete(cfg.Profiles, name)
			if err := config.Save(cfg); err != nil {
				return err
			}
			if err := mgr.Vault().DeleteToken(name); err != nil {
				return err
			}
			fmt.Printf("Perfil %s eliminado\n", name)
			return nil
		},
	}
}

func newProfileLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "Lista perfiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadCfgVault()
			if err != nil {
				return err
			}
			profiles := config.SortedProfiles(cfg)
			fmt.Println("NAME\tACCOUNT\tTAGS\tALIASES\tFAVORITE")
			for _, p := range profiles {
				fmt.Printf("%s\t%s\t%s\t%s\t%t\n", p.Name, p.AccountLabel, strings.Join(p.Tags, ","), strings.Join(p.Aliases, ","), p.Favorite)
			}
			return nil
		},
	}
}

func profileToken(input string) (string, string, *model.Config, error) {
	cfg, mgr, err := loadCfgVault()
	if err != nil {
		return "", "", nil, err
	}
	name, _, err := config.ResolveProfile(cfg, input)
	if err != nil {
		return "", "", nil, err
	}
	tk, err := mgr.Vault().GetToken(name)
	if err != nil {
		return "", "", nil, err
	}
	return name, tk, cfg, nil
}

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <profile> -- <command ...>",
		Short: "Ejecuta comando con perfil aislado",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dash := cmd.ArgsLenAtDash()
			if len(args) < 2 || dash < 0 || dash >= len(args) {
				return errors.New("usa: supashift run <profile> -- <comando>")
			}
			profile := args[0]
			command := args[dash:]
			name, token, cfg, err := profileToken(profile)
			if err != nil {
				return err
			}
			if err := runner.Run(name, token, command); err != nil {
				return err
			}
			config.TouchRecent(cfg, name)
			return config.Save(cfg)
		},
	}
}

func newShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell <profile>",
		Short: "Abre subshell con perfil aislado",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, token, cfg, err := profileToken(args[0])
			if err != nil {
				return err
			}
			if err := runner.Shell(name, token); err != nil {
				return err
			}
			config.TouchRecent(cfg, name)
			return config.Save(cfg)
		},
	}
}

func chooseManyProfiles(cfg *model.Config) ([]string, error) {
	keys := make([]string, 0, len(cfg.Profiles))
	for k := range cfg.Profiles {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println("Perfiles disponibles:")
	for i, k := range keys {
		fmt.Printf("[%d] %s\n", i+1, k)
	}
	fmt.Print("Selecciona índices separados por coma: ")
	in := bufio.NewReader(os.Stdin)
	line, _ := in.ReadString('\n')
	parts := strings.Split(strings.TrimSpace(line), ",")
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var idx int
		_, err := fmt.Sscanf(p, "%d", &idx)
		if err != nil || idx < 1 || idx > len(keys) {
			return nil, fmt.Errorf("índice inválido: %s", p)
		}
		out = append(out, keys[idx-1])
	}
	if len(out) == 0 {
		return nil, errors.New("sin perfiles seleccionados")
	}
	return out, nil
}

func newTmuxCmd() *cobra.Command {
	var many bool
	cmd := &cobra.Command{
		Use:   "tmux <profile>",
		Short: "Abre/adjunta sesión tmux aislada por perfil",
		Args: func(cmd *cobra.Command, args []string) error {
			if many {
				return nil
			}
			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, mgr, err := loadCfgVault()
			if err != nil {
				return err
			}
			if _, err := exec.LookPath("tmux"); err != nil {
				return errors.New("tmux no encontrado en PATH")
			}
			if many {
				selected, err := chooseManyProfiles(cfg)
				if err != nil {
					return err
				}
				for _, name := range selected {
					tk, err := mgr.Vault().GetToken(name)
					if err != nil {
						return err
					}
					if err := runner.TmuxDetached(name, tk); err != nil {
						return err
					}
					config.TouchRecent(cfg, name)
				}
				fmt.Printf("Sesiones tmux creadas: %s\n", strings.Join(selected, ", "))
				return config.Save(cfg)
			}
			name, _, err := config.ResolveProfile(cfg, args[0])
			if err != nil {
				return err
			}
			tk, err := mgr.Vault().GetToken(name)
			if err != nil {
				return err
			}
			if err := runner.Tmux(name, tk); err != nil {
				return err
			}
			config.TouchRecent(cfg, name)
			return config.Save(cfg)
		},
	}
	cmd.Flags().BoolVar(&many, "many", false, "Selecciona varios perfiles")
	return cmd
}

func newPickCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pick",
		Short: "Selector interactivo rápido",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, mgr, err := loadCfgVault()
			if err != nil {
				return err
			}
			choice, err := tui.PickProfile(cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Acción [shell|tmux|use|run]: ")
			in := bufio.NewReader(os.Stdin)
			act, _ := in.ReadString('\n')
			act = strings.TrimSpace(act)
			name, _, err := config.ResolveProfile(cfg, choice)
			if err != nil {
				return err
			}
			tk, err := mgr.Vault().GetToken(name)
			if err != nil {
				return err
			}
			switch act {
			case "", "shell":
				err = runner.Shell(name, tk)
			case "tmux":
				err = runner.Tmux(name, tk)
			case "use":
				fmt.Print(runner.UseSnippet(name, tk))
			case "run":
				fmt.Fprintf(os.Stderr, "Comando: ")
				line, _ := in.ReadString('\n')
				parts := strings.Fields(strings.TrimSpace(line))
				err = runner.Run(name, tk, parts)
			default:
				return fmt.Errorf("acción no soportada: %s", act)
			}
			if err != nil {
				return err
			}
			config.TouchRecent(cfg, name)
			return config.Save(cfg)
		},
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnóstico de seguridad y entorno",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, mgr, err := loadCfgVault()
			if err != nil {
				return err
			}
			path, _ := config.ConfigPath()
			fmt.Printf("config: %s\n", path)
			fmt.Printf("vault seleccionado: %s\n", mgr.Vault().Backend())
			kr, msg := mgr.Doctor()
			fmt.Printf("keyring: %v (%s)\n", kr, msg)
			if _, err := exec.LookPath("tmux"); err == nil {
				fmt.Println("tmux: OK")
			} else {
				fmt.Println("tmux: no encontrado")
			}
			fmt.Printf("profiles: %d\n", len(cfg.Profiles))
			fmt.Println("recomendación zsh: setopt HIST_IGNORE_SPACE")
			return nil
		},
	}
}

func newExportCmd() *cobra.Command {
	var output string
	var includeSecrets bool
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Exporta configuración (sin tokens por defecto)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, mgr, err := loadCfgVault()
			if err != nil {
				return err
			}
			bundle := model.ExportBundle{Version: cfg.Version, Profiles: cfg.Profiles, ProjectMappings: cfg.ProjectMappings}
			if includeSecrets {
				bundle.Tokens = map[string]string{}
				for name := range cfg.Profiles {
					tk, err := mgr.Vault().GetToken(name)
					if err != nil {
						return fmt.Errorf("no se pudo exportar secreto de %s: %w", name, err)
					}
					bundle.Tokens[name] = tk
				}
			}
			b, err := toml.Marshal(bundle)
			if err != nil {
				return err
			}
			if output == "" || output == "-" {
				fmt.Print(string(b))
				return nil
			}
			return os.WriteFile(output, b, 0o600)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "-", "Archivo de salida")
	cmd.Flags().BoolVar(&includeSecrets, "include-secrets", false, "Incluye tokens explícitamente")
	return cmd
}

func newImportCmd() *cobra.Command {
	var input string
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Importa configuración exportada",
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return errors.New("usa --input <archivo>")
			}
			cfg, mgr, err := loadCfgVault()
			if err != nil {
				return err
			}
			b, err := os.ReadFile(input)
			if err != nil {
				return err
			}
			var bundle model.ExportBundle
			if err := toml.Unmarshal(b, &bundle); err != nil {
				return err
			}
			for name, p := range bundle.Profiles {
				if p.Name == "" {
					p.Name = name
				}
				config.UpsertProfile(cfg, p)
			}
			for k, v := range bundle.ProjectMappings {
				cfg.ProjectMappings[k] = v
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			for k, v := range bundle.Tokens {
				if err := mgr.Vault().SetToken(k, v); err != nil {
					return err
				}
			}
			fmt.Printf("Importados %d perfiles\n", len(bundle.Profiles))
			return nil
		},
	}
	cmd.Flags().StringVarP(&input, "input", "i", "", "Archivo de importación")
	return cmd
}

func readLegacySupabaseToken(source, path, envName string) (string, string, error) {
	trim := func(v string) string { return strings.TrimSpace(v) }
	switch source {
	case "auto":
		if b, err := os.ReadFile(path); err == nil {
			if v := trim(string(b)); v != "" {
				return v, "file", nil
			}
		}
		if v := trim(os.Getenv(envName)); v != "" {
			return v, "env", nil
		}
		return "", "", fmt.Errorf("no se encontró token en %s ni en $%s", path, envName)
	case "file":
		b, err := os.ReadFile(path)
		if err != nil {
			return "", "", err
		}
		v := trim(string(b))
		if v == "" {
			return "", "", errors.New("token vacío en archivo")
		}
		return v, "file", nil
	case "env":
		v := trim(os.Getenv(envName))
		if v == "" {
			return "", "", fmt.Errorf("variable %s vacía o inexistente", envName)
		}
		return v, "env", nil
	case "stdin":
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", "", err
		}
		v := trim(string(b))
		if v == "" {
			return "", "", errors.New("stdin sin token")
		}
		return v, "stdin", nil
	default:
		return "", "", fmt.Errorf("source no soportado: %s", source)
	}
}

func newMigrateCmd() *cobra.Command {
	var account, notes, tags, aliases, source, fromFile, envName string
	var favorite bool
	cmd := &cobra.Command{
		Use:   "migrate-from-supabase-cli <profile>",
		Short: "Migra token existente de Supabase CLI a un perfil seguro",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg, mgr, err := loadCfgVault()
			if err != nil {
				return err
			}
			token, usedSource, err := readLegacySupabaseToken(source, fromFile, envName)
			if err != nil {
				return err
			}
			p, exists := cfg.Profiles[name]
			if !exists {
				p = model.Profile{Name: name}
			}
			if cmd.Flags().Changed("account") {
				p.AccountLabel = account
			}
			if p.AccountLabel == "" {
				p.AccountLabel = name
			}
			if cmd.Flags().Changed("notes") {
				p.Notes = notes
			}
			if cmd.Flags().Changed("tags") {
				p.Tags = parseCSV(tags)
			}
			if cmd.Flags().Changed("aliases") {
				p.Aliases = parseCSV(aliases)
			}
			if cmd.Flags().Changed("favorite") {
				p.Favorite = favorite
			}
			config.UpsertProfile(cfg, p)
			if err := config.Save(cfg); err != nil {
				return err
			}
			if err := mgr.Vault().SetToken(name, token); err != nil {
				return err
			}
			fmt.Printf("Perfil %s migrado desde %s (backend=%s)\n", name, usedSource, mgr.Vault().Backend())
			return nil
		},
	}
	home, _ := os.UserHomeDir()
	cmd.Flags().StringVar(&source, "source", "auto", "Fuente token: auto|file|env|stdin")
	cmd.Flags().StringVar(&fromFile, "from-file", filepath.Join(home, ".supabase", "access-token"), "Ruta de token legacy")
	cmd.Flags().StringVar(&envName, "env", "SUPABASE_ACCESS_TOKEN", "Variable de entorno origen")
	cmd.Flags().StringVar(&account, "account", "", "Etiqueta de cuenta")
	cmd.Flags().StringVar(&notes, "notes", "", "Notas")
	cmd.Flags().StringVar(&tags, "tags", "", "Tags separadas por coma")
	cmd.Flags().StringVar(&aliases, "aliases", "", "Aliases separadas por coma")
	cmd.Flags().BoolVar(&favorite, "favorite", false, "Marcar favorito")
	return cmd
}

func newUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Imprime snippet para eval",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, tk, cfg, err := profileToken(args[0])
			if err != nil {
				return err
			}
			fmt.Print(runner.UseSnippet(name, tk))
			config.TouchRecent(cfg, name)
			return config.Save(cfg)
		},
	}
}

func newUnuseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unuse",
		Short: "Imprime snippet para limpiar entorno",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(runner.UnuseSnippet())
			return nil
		},
	}
}

func newAutoCmd() *cobra.Command {
	var set bool
	cmd := &cobra.Command{
		Use:   "auto",
		Short: "Sugiere perfil según carpeta actual",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadCfgVault()
			if err != nil {
				return err
			}
			cwd, _ := os.Getwd()
			root, ok := integrations.DetectSupabaseProject(cwd)
			if !ok {
				fmt.Println("No se detectó proyecto Supabase")
				return nil
			}
			profile := cfg.ProjectMappings[root]
			if profile == "" {
				fmt.Printf("Proyecto detectado en %s, sin mapping\n", root)
				return nil
			}
			if set {
				fmt.Print(autoSetSnippet(profile))
				return nil
			}
			fmt.Printf("Perfil sugerido: %s (proyecto %s)\n", profile, root)
			return nil
		},
	}
	cmd.Flags().BoolVar(&set, "set", false, "Imprime comando use listo para eval")
	return cmd
}

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Integración con repositorios"}
	var path string
	bind := &cobra.Command{
		Use:   "bind <profile>",
		Short: "Asocia perfil al proyecto actual o a --path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadCfgVault()
			if err != nil {
				return err
			}
			name, _, err := config.ResolveProfile(cfg, args[0])
			if err != nil {
				return err
			}
			target := path
			if strings.TrimSpace(target) == "" {
				cwd, _ := os.Getwd()
				root, ok := integrations.DetectSupabaseProject(cwd)
				if !ok {
					return errors.New("no se detectó proyecto Supabase; usa --path")
				}
				target = root
			}
			abs, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			cfg.ProjectMappings[abs] = name
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Printf("Mapping guardado: %s -> %s\n", abs, name)
			return nil
		},
	}
	bind.Flags().StringVar(&path, "path", "", "Ruta del proyecto")
	ls := &cobra.Command{
		Use:   "ls",
		Short: "Lista mappings de proyectos",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadCfgVault()
			if err != nil {
				return err
			}
			keys := make([]string, 0, len(cfg.ProjectMappings))
			for k := range cfg.ProjectMappings {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%s\t%s\n", k, cfg.ProjectMappings[k])
			}
			return nil
		},
	}
	cmd.AddCommand(bind, ls)
	return cmd
}

func newRevealCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reveal <profile>",
		Short: "Revela token (requiere confirmación explícita)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !vault.ConfirmReveal() {
				return errors.New("confirmación fallida")
			}
			_, tk, _, err := profileToken(args[0])
			if err != nil {
				return err
			}
			fmt.Println(tk)
			return nil
		},
	}
}

func newCompletionsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "completions", Short: "Genera completions"}
	z := &cobra.Command{
		Use:   "zsh",
		Short: "Completion zsh",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(os.Stdout)
		},
	}
	cmd.AddCommand(z)
	return cmd
}

func newSnippetCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "snippet", Short: "Snippets para shell"}
	zsh := &cobra.Command{
		Use:   "zsh",
		Short: "Snippet de integración zsh",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = io.WriteString(os.Stdout, `
# Supashift Zsh integration
setopt HIST_IGNORE_SPACE
supause() { eval "$(supashift use "$1")"; }
supaunuse() { eval "$(supashift unuse)"; }
# Auto-suggest profile in Supabase repos:
autoload -Uz add-zsh-hook
supashift_auto() {
  local suggestion
  suggestion="$(supashift auto 2>/dev/null | head -n1)"
  [[ -n "$suggestion" ]] && print -P "%F{242}${suggestion}%f"
}
add-zsh-hook chpwd supashift_auto
`)
		},
	}
	cmd.AddCommand(zsh)
	return cmd
}

func BindProject(path, profile string) error {
	cfg, _, err := loadCfgVault()
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	cfg.ProjectMappings[abs] = profile
	return config.Save(cfg)
}
