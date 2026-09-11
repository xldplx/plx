package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Inspect inspects a git repository directory and extracts its current status.
func Inspect(repoPath string) (*Repo, error) {
	name := filepath.Base(repoPath)

	repo := &Repo{
		Name:          name,
		Path:          repoPath,
		Branch:        "unknown",
		IsClean:       true,
		RecentCommits: []string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Get branch
	branchCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if out, err := branchCmd.Output(); err == nil {
		branch := strings.TrimSpace(string(out))
		if branch == "HEAD" {
			// Detached HEAD, get short hash
			hashCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "--short", "HEAD")
			if hOut, hErr := hashCmd.Output(); hErr == nil {
				repo.Branch = ":" + strings.TrimSpace(string(hOut))
			} else {
				repo.Branch = "detached"
			}
		} else {
			repo.Branch = branch
		}
	}

	// 2. Get status --porcelain
	statusCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "status", "--porcelain=v1")
	if out, err := statusCmd.Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		var changed []string
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			if len(line) < 3 {
				continue
			}
			x := line[0]
			y := line[1]
			file := strings.TrimSpace(line[3:])

			if x == '?' && y == '?' {
				repo.UntrackedCount++
			} else {
				if x != ' ' && x != '?' {
					repo.StagedCount++
				}
				if y != ' ' {
					repo.ModifiedCount++
				}
			}

			if len(changed) < 5 && file != "" {
				prefix := ""
				if x == '?' && y == '?' {
					prefix = "[?]"
				} else if x != ' ' {
					prefix = "[+]"
				} else {
					prefix = "[M]"
				}
				changed = append(changed, fmt.Sprintf("%s %s", prefix, file))
			}
		}
		repo.ChangedFiles = changed

		totalDirty := repo.ModifiedCount + repo.UntrackedCount + repo.StagedCount
		if totalDirty > 0 {
			repo.IsClean = false
			var parts []string
			if repo.ModifiedCount > 0 {
				parts = append(parts, fmt.Sprintf("%d modified", repo.ModifiedCount))
			}
			if repo.StagedCount > 0 {
				parts = append(parts, fmt.Sprintf("%d staged", repo.StagedCount))
			}
			if repo.UntrackedCount > 0 {
				parts = append(parts, fmt.Sprintf("%d untracked", repo.UntrackedCount))
			}
			repo.DirtySummary = strings.Join(parts, ", ")
		} else {
			repo.DirtySummary = "clean"
		}
	}

	// 3. Ahead / Behind upstream
	aheadCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if out, err := aheadCmd.Output(); err == nil {
		fields := strings.Fields(string(out))
		if len(fields) >= 2 {
			repo.Ahead, _ = strconv.Atoi(fields[0])
			repo.Behind, _ = strconv.Atoi(fields[1])
		}
	}

	// 4. Recent commits (last 3)
	logCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "log", "-3", "--format=%h%x00%cr%x00%s")
	if out, err := logCmd.Output(); err == nil {
		lines := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
		for _, line := range lines {
			line = bytes.TrimRight(line, "\r")
			parts := bytes.Split(line, []byte{0})
			if len(parts) >= 3 {
				hash := string(parts[0])
				age := string(parts[1])
				msg := string(parts[2])
				repo.RecentCommits = append(repo.RecentCommits, fmt.Sprintf("%s · %s · %s", hash, age, msg))
				if repo.LastCommitAge == "" {
					repo.LastCommitAge = age
					repo.LastCommitMsg = msg
				}
			}
		}
	}

	return repo, nil
}