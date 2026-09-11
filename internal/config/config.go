package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config defines the user settings for plx.
type Config struct {
	WorkspaceRoots []string `toml:"workspace_roots"`
	MaxDepth       int      `toml:"max_depth"`
	IgnoreDirs     []string `toml:"ignore_dirs"`
	DefaultEditor  string   `toml:"default_editor"`
}

// DefaultConfig generates a sane default configuration.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	roots := []string{}
	if home != "" {
		// Include common project discovery paths if they exist
		candidates := []string{
			filepath.Join(home, "Desktop"),
			filepath.Join(home, "Projects"),
			filepath.Join(home, "src"),
			filepath.Join(home, "code"),
			filepath.Join(home, "work"),
			filepath.Join(home, "dev"),
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && info.IsDir() {
				roots = append(roots, c)
			}
		}
		if len(roots) == 0 {
			roots = append(roots, home)
		}
	}

	return &Config{
		WorkspaceRoots: roots,
		MaxDepth:       4,
		IgnoreDirs: []string{
			"node_modules",
			"vendor",
			"target",
			".cargo",
			"dist",
			".next",
			".venv",
			"venv",
			"__pycache__",
			"build",
			"bin",
			"obj",
			".git",
		},
		DefaultEditor: "code",
	}
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				return home
			}
			if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
				return filepath.Join(home, p[2:])
			}
		}
	}
	return p
}

// ConfigFilePath returns the full path to config.toml.
func ConfigFilePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// LoadConfig reads config.toml from the XDG config directory, creating a default one if missing.
func LoadConfig() (*Config, error) {
	cfgFile, err := ConfigFilePath()
	if err != nil {
		return DefaultConfig(), nil
	}

	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		cfg := DefaultConfig()
		_ = SaveConfig(cfg)
		return cfg, nil
	}

	var cfg Config
	if _, err := toml.DecodeFile(cfgFile, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to parse config file %s: %w", cfgFile, err)
	}

	if len(cfg.WorkspaceRoots) == 0 {
		cfg.WorkspaceRoots = DefaultConfig().WorkspaceRoots
	}
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 4
	}
	if len(cfg.IgnoreDirs) == 0 {
		cfg.IgnoreDirs = DefaultConfig().IgnoreDirs
	}

	return &cfg, nil
}

// SaveConfig writes the configuration out to config.toml atomically.
func SaveConfig(cfg *Config) error {
	cfgFile, err := ConfigFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cfgFile), 0755); err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(cfg); err != nil {
		return err
	}

	tmpFile := fmt.Sprintf("%s.tmp.%d", cfgFile, os.Getpid())
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, cfgFile)
}
