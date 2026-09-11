package git

// Repo holds the parsed state and Git metadata for a local repository.
type Repo struct {
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	Branch         string   `json:"branch"`
	IsClean        bool     `json:"is_clean"`
	ModifiedCount  int      `json:"modified_count"`
	UntrackedCount int      `json:"untracked_count"`
	StagedCount    int      `json:"staged_count"`
	Ahead          int      `json:"ahead"`
	Behind         int      `json:"behind"`
	LastCommitAge  string   `json:"last_commit_age"`
	LastCommitMsg  string   `json:"last_commit_msg"`
	DirtySummary   string   `json:"dirty_summary"`
	ChangedFiles   []string `json:"changed_files"`
	RecentCommits  []string `json:"recent_commits"`
}