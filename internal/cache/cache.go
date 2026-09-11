package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xldplx/plx/internal/config"
	"github.com/xldplx/plx/internal/git"
)

// StateCache holds persisted repository snapshots.
type StateCache struct {
	LastScan time.Time   `json:"last_scan"`
	Repos    []*git.Repo `json:"repos"`
}

// CacheFilePath returns the path to cache.json in the state directory.
func CacheFilePath() (string, error) {
	stateDir, err := config.GetStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "cache.json"), nil
}

// Load reads the cache file. If not found or corrupted, returns an empty cache.
func Load() (*StateCache, error) {
	cachePath, err := CacheFilePath()
	if err != nil {
		return &StateCache{Repos: []*git.Repo{}}, nil
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return &StateCache{Repos: []*git.Repo{}}, nil
	}

	var sc StateCache
	if err := json.Unmarshal(data, &sc); err != nil {
		return &StateCache{Repos: []*git.Repo{}}, nil
	}

	return &sc, nil
}

// Save writes the cache atomically via a temporary file.
func Save(repos []*git.Repo) error {
	cachePath, err := CacheFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	sc := StateCache{
		LastScan: time.Now(),
		Repos:    repos,
	}

	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := fmt.Sprintf("%s.tmp.%d", cachePath, os.Getpid())
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, cachePath)
}
