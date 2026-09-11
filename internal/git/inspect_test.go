package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInspect(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "plx-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		// Try fallback without -b main
		cmd2 := exec.Command("git", "init")
		cmd2.Dir = tempDir
		if err := cmd2.Run(); err != nil {
			t.Skip("git not available or failed to init repo")
		}
	}

	// Create an untracked file
	testFile := filepath.Join(tempDir, "sample.txt")
	if err := os.WriteFile(testFile, []byte("hello plx"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	repo, err := Inspect(tempDir)
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}

	if repo.IsClean {
		t.Errorf("expected repo to be dirty due to untracked file")
	}

	if repo.UntrackedCount < 1 {
		t.Errorf("expected untracked count >= 1, got %d", repo.UntrackedCount)
	}
}
