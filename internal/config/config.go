package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/villawebcl/supashift/internal/model"
)

const (
	AppName    = "supashift"
	ConfigFile = "config.toml"
)

func ConfigDir() (string, error) {
	base, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok || strings.TrimSpace(base) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, AppName), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFile), nil
}

func EnsureDirs() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func defaultConfig() *model.Config {
	return &model.Config{
		Version:         1,
		VaultBackend:    "auto",
		Profiles:        map[string]model.Profile{},
		Recents:         []string{},
		ProjectMappings: map[string]string{},
	}
}

func InitConfig() (*model.Config, string, error) {
	dir, err := EnsureDirs()
	if err != nil {
		return nil, "", err
	}
	path := filepath.Join(dir, ConfigFile)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		cfg := defaultConfig()
		if err := Save(cfg); err != nil {
			return nil, "", err
		}
		return cfg, path, nil
	}
	cfg, err := Load()
	return cfg, path, err
}

func Load() (*model.Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultConfig(), nil
	}
	if err != nil {
		return nil, err
	}
	cfg := defaultConfig()
	if err := toml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("config inválida: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]model.Profile{}
	}
	if cfg.ProjectMappings == nil {
		cfg.ProjectMappings = map[string]string{}
	}
	return cfg, nil
}

func Save(cfg *model.Config) error {
	if _, err := EnsureDirs(); err != nil {
		return err
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	b, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	return nil
}

func ResolveProfile(cfg *model.Config, input string) (string, model.Profile, error) {
	if p, ok := cfg.Profiles[input]; ok {
		return input, p, nil
	}
	for name, p := range cfg.Profiles {
		for _, a := range p.Aliases {
			if a == input {
				return name, p, nil
			}
		}
	}
	return "", model.Profile{}, fmt.Errorf("perfil no encontrado: %s", input)
}

func TouchRecent(cfg *model.Config, profile string) {
	out := []string{profile}
	for _, v := range cfg.Recents {
		if v != profile {
			out = append(out, v)
		}
		if len(out) >= 20 {
			break
		}
	}
	cfg.Recents = out
}

func SortedProfiles(cfg *model.Config) []model.Profile {
	items := make([]model.Profile, 0, len(cfg.Profiles))
	for _, p := range cfg.Profiles {
		items = append(items, p)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Favorite != items[j].Favorite {
			return items[i].Favorite
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	return items
}

func UpsertProfile(cfg *model.Config, p model.Profile) {
	p.UpdatedAt = time.Now().UTC()
	cfg.Profiles[p.Name] = p
}
