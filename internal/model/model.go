package model

import "time"

type Profile struct {
	Name         string    `toml:"name" json:"name"`
	AccountLabel string    `toml:"account_label" json:"account_label"`
	Notes        string    `toml:"notes,omitempty" json:"notes,omitempty"`
	Tags         []string  `toml:"tags,omitempty" json:"tags,omitempty"`
	Aliases      []string  `toml:"aliases,omitempty" json:"aliases,omitempty"`
	Favorite     bool      `toml:"favorite,omitempty" json:"favorite,omitempty"`
	UpdatedAt    time.Time `toml:"updated_at" json:"updated_at"`
}

type Config struct {
	Version         int                `toml:"version" json:"version"`
	VaultBackend    string             `toml:"vault_backend,omitempty" json:"vault_backend,omitempty"`
	Profiles        map[string]Profile `toml:"profiles" json:"profiles"`
	Recents         []string           `toml:"recents,omitempty" json:"recents,omitempty"`
	ProjectMappings map[string]string  `toml:"project_mappings,omitempty" json:"project_mappings,omitempty"`
}

type ExportBundle struct {
	Version         int                `toml:"version" json:"version"`
	Profiles        map[string]Profile `toml:"profiles" json:"profiles"`
	ProjectMappings map[string]string  `toml:"project_mappings,omitempty" json:"project_mappings,omitempty"`
	Tokens          map[string]string  `toml:"tokens,omitempty" json:"tokens,omitempty"`
}
