package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the persisted user configuration.
type Config struct {
	APIKey     string   `json:"api_key"`
	Workspaces []string `json:"workspaces"` // allowed workspace IDs — hard lock
	SecureMode bool     `json:"secure_mode"` // when true, only read + create operations allowed
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "typeform", "config.json"), nil
}

// Load reads the config file. Returns empty Config (not error) if file doesn't exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config file with 0600 permissions.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Clear removes the config file (logout).
func Clear() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Path returns the config file path for display purposes.
func Path() string {
	p, _ := configPath()
	return p
}

// IsWorkspaceAllowed checks if a workspace ID is in the allowed list.
// If no workspaces are configured, returns an error (must configure at least one).
func IsWorkspaceAllowed(cfg *Config, workspaceID string) error {
	if len(cfg.Workspaces) == 0 {
		return fmt.Errorf("no workspaces configured — add workspace IDs to config file: %s", Path())
	}
	for _, w := range cfg.Workspaces {
		if w == workspaceID {
			return nil
		}
	}
	return fmt.Errorf("workspace %q is not in the allowed list — allowed: %v", workspaceID, cfg.Workspaces)
}

// CheckSecureMode returns an error if secure mode is on and the operation is not allowed.
// In secure mode, only "read" and "create" operations are permitted.
func CheckSecureMode(cfg *Config, operation string) error {
	if !cfg.SecureMode {
		return nil
	}
	switch operation {
	case "read", "create":
		return nil
	default:
		return fmt.Errorf("secure mode is ON — operation %q is blocked (only read and create allowed)\nTo disable: set \"secure_mode\": false in %s", operation, Path())
	}
}
