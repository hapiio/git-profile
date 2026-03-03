// Package config manages the git-profile configuration file.
// Profiles are stored in $XDG_CONFIG_HOME/git-profile/config.json (or
// $HOME/.config/git-profile/config.json on systems without XDG).
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const currentVersion = 1

// Profile represents a single named git identity.
type Profile struct {
	ID          string `json:"id"`
	GitUser     string `json:"git_user"`
	GitEmail    string `json:"git_email"`
	SSHKeyPath  string `json:"ssh_key_path,omitempty"`
	GPGKeyID    string `json:"gpg_key_id,omitempty"`
	SignCommits bool   `json:"sign_commits,omitempty"`
}

// Config is the root configuration structure.
type Config struct {
	Version  int                `json:"version"`
	Profiles map[string]Profile `json:"profiles"`
}

// SortedIDs returns profile IDs in deterministic alphabetical order.
func (c *Config) SortedIDs() []string {
	ids := make([]string, 0, len(c.Profiles))
	for id := range c.Profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Manager provides config load/save operations bound to a specific file path.
type Manager struct {
	path string
}

// NewManager returns a Manager. When override is non-empty it is used as the
// config file path; otherwise the platform-default path is resolved.
func NewManager(override string) (*Manager, error) {
	if override != "" {
		dir := filepath.Dir(override)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("creating config directory: %w", err)
		}
		return &Manager{path: override}, nil
	}
	p, err := defaultConfigPath()
	if err != nil {
		return nil, err
	}
	return &Manager{path: p}, nil
}

func defaultConfigPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}
	dir := filepath.Join(cfgDir, "git-profile")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("creating config directory: %w", err)
	}
	return filepath.Join(dir, "config.json"), nil
}

// Path returns the resolved config file path.
func (m *Manager) Path() string { return m.path }

// Load reads and returns the config. Returns an empty config if the file
// does not exist yet.
func (m *Manager) Load() (*Config, error) {
	cfg := &Config{
		Version:  currentVersion,
		Profiles: make(map[string]Profile),
	}

	f, err := os.Open(m.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("opening config: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", m.path, err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	// Apply migrations for older config versions.
	migrate(cfg)

	return cfg, nil
}

// Save writes cfg to disk atomically (write-to-temp then rename).
func (m *Manager) Save(cfg *Config) error {
	tmp := m.path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("creating temp config: %w", err)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if encErr := enc.Encode(cfg); encErr != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("encoding config: %w", encErr)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("flushing config: %w", err)
	}

	if err := os.Rename(tmp, m.path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}

// migrate applies forward-compatibility fixes to configs created by older versions.
func migrate(cfg *Config) {
	if cfg.Version < 1 {
		cfg.Version = 1
	}
	// Ensure profile IDs are consistent with map keys.
	for id, p := range cfg.Profiles {
		if p.ID == "" {
			p.ID = id
			cfg.Profiles[id] = p
		}
	}
}
