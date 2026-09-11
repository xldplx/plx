package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/xldplx/plx/internal/config"
	"github.com/xldplx/plx/internal/git"
)

// Scanner coordinates finding and inspecting repositories across workspace roots.
type Scanner struct {
	cfg *config.Config
}

// New creates a new Scanner instance with given config.
func New(cfg *config.Config) *Scanner {
	return &Scanner{cfg: cfg}
}

// Scan crawls all configured workspace roots and returns inspected Git repos.
func (s *Scanner) Scan() ([]*git.Repo, error) {
	repoPaths := make(map[string]bool)
	var mu sync.Mutex

	ignoreMap := make(map[string]bool)
	for _, ign := range s.cfg.IgnoreDirs {
		ignoreMap[strings.ToLower(ign)] = true
	}

	for _, rawRoot := range s.cfg.WorkspaceRoots {
		root := config.ExpandPath(rawRoot)
		if root == "" {
			continue
		}
		if _, err := os.Stat(root); err != nil {
			continue
		}

		s.walkDir(root, 0, ignoreMap, func(path string) {
			mu.Lock()
			repoPaths[path] = true
			mu.Unlock()
		})
	}

	// Concurrently inspect all discovered repos
	paths := make([]string, 0, len(repoPaths))
	for p := range repoPaths {
		paths = append(paths, p)
	}

	results := make([]*git.Repo, 0, len(paths))
	var resMu sync.Mutex

	concurrency := 16
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, p := range paths {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(repoPath string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			r, err := git.Inspect(repoPath)
			if err == nil && r != nil {
				resMu.Lock()
				results = append(results, r)
				resMu.Unlock()
			}
		}(p)
	}
	wg.Wait()

	// Sort: Dirty repositories first, then alphabetically by Name
	sort.Slice(results, func(i, j int) bool {
		if results[i].IsClean != results[j].IsClean {
			return !results[i].IsClean // dirty first
		}
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	return results, nil
}

// walkDir recursively traverses directories respecting depth and ignores.
func (s *Scanner) walkDir(dir string, currentDepth int, ignoreMap map[string]bool, onRepoFound func(path string)) {
	if currentDepth > s.cfg.MaxDepth {
		return
	}

	// Check if this dir is a git repo
	gitPath := filepath.Join(dir, ".git")
	if fi, err := os.Stat(gitPath); err == nil && (fi.IsDir() || fi.Mode().IsRegular()) {
		onRepoFound(dir)
		return // Do not recurse inside a git repo
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Ignore hidden directories (except we checked .git) and ignored names
		if strings.HasPrefix(name, ".") || ignoreMap[strings.ToLower(name)] {
			continue
		}

		subDir := filepath.Join(dir, name)
		s.walkDir(subDir, currentDepth+1, ignoreMap, onRepoFound)
	}
}
