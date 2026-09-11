package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
	"github.com/xldplx/plx/internal/cache"
	"github.com/xldplx/plx/internal/config"
	"github.com/xldplx/plx/internal/git"
	"github.com/xldplx/plx/internal/scanner"
	"github.com/xldplx/plx/internal/tui"
)

var (
	version    = "dev"
	commit     = "none"
	date       = "unknown"
	jsonOutput bool
	dirtyOnly  bool
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	rootCmd := &cobra.Command{
		Use:     "plx",
		Short:   "plx - Lightweight local-first workspace flight deck by @xldplx",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		RunE: func(cmd *cobra.Command, args []string) error {
			// If not a TTY or explicitly piped, run list
			fi, _ := os.Stdout.Stat()
			isPipe := (fi.Mode() & os.ModeCharDevice) == 0

			if isPipe {
				return runList(cfg)
			}

			// Launch interactive TUI
			m := tui.New(cfg)
			p := tea.NewProgram(m, tea.WithAltScreen())
			finalModel, err := p.Run()
			if err != nil {
				return err
			}

			if tm, ok := finalModel.(tui.Model); ok && tm.SelectedPath != "" {
				// Print selected path cleanly so shell wrapper can cd
				fmt.Println(tm.SelectedPath)
			}
			return nil
		},
	}

	// Subcommands
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered workspace repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cfg)
		},
	}
	listCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	listCmd.Flags().BoolVar(&dirtyOnly, "dirty", false, "Only list repositories with uncommitted changes")

	jumpCmd := &cobra.Command{
		Use:   "jump [query]",
		Short: "Resolve repository path for rapid directory hopping",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJump(cfg, args[0])
		},
	}

	scanCmd := &cobra.Command{
		Use:   "scan [path]",
		Short: "Rescan workspace roots and refresh local cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(cfg, args)
		},
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Display active configuration and paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgFile, _ := config.ConfigFilePath()
			cacheFile, _ := cache.CacheFilePath()
			fmt.Printf("Config File: %s\n", cfgFile)
			fmt.Printf("Cache File:  %s\n\n", cacheFile)
			fmt.Printf("Workspace Roots:\n")
			for _, r := range cfg.WorkspaceRoots {
				fmt.Printf("  • %s\n", config.ExpandPath(r))
			}
			fmt.Printf("\nMax Depth:      %d\n", cfg.MaxDepth)
			fmt.Printf("Default Editor: %s\n", cfg.DefaultEditor)
			return nil
		},
	}

	configEditCmd := &cobra.Command{
		Use:   "edit",
		Short: "Open config.toml in your default editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgFile, err := config.ConfigFilePath()
			if err != nil {
				return err
			}
			if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
				_ = config.SaveConfig(cfg)
			}
			editor := cfg.DefaultEditor
			if editor == "" {
				editor = "code"
			}
			c := exec.Command(editor, cfgFile)
			if runtime.GOOS == "windows" && editor == "code" {
				c = exec.Command("cmd", "/c", "code", cfgFile)
			}
			return c.Start()
		},
	}
	configCmd.AddCommand(configEditCmd)

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Automatically install the shell jump wrappers into your shell profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}

	initCmd := &cobra.Command{
		Use:   "init [powershell|bash|zsh|fish]",
		Short: "Print shell integration wrapper functions for rapid jumping",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shell := "powershell"
			if len(args) > 0 {
				shell = strings.ToLower(args[0])
			}
			printShellInit(shell)
		},
	}

	rootCmd.AddCommand(listCmd, jumpCmd, scanCmd, configCmd, setupCmd, initCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getRepos(cfg *config.Config) ([]*git.Repo, error) {
	cached, _ := cache.Load()
	if cached != nil && len(cached.Repos) > 0 {
		return cached.Repos, nil
	}

	s := scanner.New(cfg)
	repos, err := s.Scan()
	if err != nil {
		return nil, err
	}
	_ = cache.Save(repos)
	return repos, nil
}

func runList(cfg *config.Config) error {
	repos, err := getRepos(cfg)
	if err != nil {
		return err
	}

	var filtered []*git.Repo
	for _, r := range repos {
		if dirtyOnly && r.IsClean {
			continue
		}
		filtered = append(filtered, r)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(filtered)
	}

	fmt.Printf("%-24s %-16s %-24s %-8s %s\n", "NAME", "BRANCH", "STATUS", "SYNC", "PATH")
	fmt.Println(strings.Repeat("-", 90))
	for _, r := range filtered {
		sync := fmt.Sprintf("↑%d ↓%d", r.Ahead, r.Behind)
		fmt.Printf("%-24s %-16s %-24s %-8s %s\n",
			truncate(r.Name, 24),
			truncate(r.Branch, 16),
			truncate(r.DirtySummary, 24),
			sync,
			r.Path,
		)
	}
	return nil
}

func runJump(cfg *config.Config, query string) error {
	repos, err := getRepos(cfg)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories discovered")
	}

	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.Name + " " + r.Path
	}

	matches := fuzzy.Find(query, names)
	if len(matches) == 0 {
		return fmt.Errorf("no repository matched query: %s", query)
	}

	best := repos[matches[0].Index]
	fmt.Println(best.Path)
	return nil
}

func runScan(cfg *config.Config, args []string) error {
	if len(args) > 0 {
		cfg.WorkspaceRoots = append(cfg.WorkspaceRoots, args[0])
		_ = config.SaveConfig(cfg)
	}

	fmt.Println("Scanning workspace roots...")
	s := scanner.New(cfg)
	repos, err := s.Scan()
	if err != nil {
		return err
	}

	if err := cache.Save(repos); err != nil {
		return err
	}

	dirtyCount := 0
	for _, r := range repos {
		if !r.IsClean {
			dirtyCount++
		}
	}

	fmt.Printf("Scan complete. Discovered %d repositories (%d dirty).\n", len(repos), dirtyCount)
	return nil
}

func runSetup() error {
	if runtime.GOOS == "windows" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		psDir := filepath.Join(home, "Documents", "WindowsPowerShell")
		profilePath := filepath.Join(psDir, "Microsoft.PowerShell_profile.ps1")

		if err := os.MkdirAll(psDir, 0755); err != nil {
			return err
		}

		hookSnippet := "\n# plx shell integration (https://github.com/xldplx/plx)\nplx init powershell | Out-String | Invoke-Expression\n"

		if data, err := os.ReadFile(profilePath); err == nil {
			if strings.Contains(string(data), "plx init powershell") {
				fmt.Printf("✓ plx is already configured in your PowerShell profile:\n  %s\n\nRun '. $PROFILE' or open a new terminal to use 'plx' or 'x'!\n", profilePath)
				return nil
			}
		}

		f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open profile: %w", err)
		}
		defer f.Close()

		if _, err := f.WriteString(hookSnippet); err != nil {
			return fmt.Errorf("failed to write hook to profile: %w", err)
		}

		fmt.Printf("✓ Successfully configured plx in your PowerShell profile:\n  %s\n\nTo activate it in this terminal now, run:\n  . $PROFILE\n\nThen type 'plx' or 'x' to jump into your projects!\n", profilePath)
		return nil
	}

	// Unix (macOS / Linux)
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	shell := os.Getenv("SHELL")
	targetFile := filepath.Join(home, ".bashrc")
	shellName := "bash"
	if strings.Contains(shell, "zsh") {
		targetFile = filepath.Join(home, ".zshrc")
		shellName = "zsh"
	}

	hookSnippet := fmt.Sprintf("\n# plx shell integration (https://github.com/xldplx/plx)\neval \"$(plx init %s)\"\n", shellName)

	if data, err := os.ReadFile(targetFile); err == nil {
		if strings.Contains(string(data), "plx init") {
			fmt.Printf("✓ plx is already configured in:\n  %s\n\nRun 'source %s' or open a new terminal to use 'plx' or 'x'!\n", targetFile, targetFile)
			return nil
		}
	}

	f, err := os.OpenFile(targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", targetFile, err)
	}
	defer f.Close()

	if _, err := f.WriteString(hookSnippet); err != nil {
		return fmt.Errorf("failed to write to %s: %w", targetFile, err)
	}

	fmt.Printf("✓ Successfully configured plx in:\n  %s\n\nTo activate it now, run:\n  source %s\n\nThen type 'plx' or 'x' to jump into your projects!\n", targetFile, targetFile)
	return nil
}

func printShellInit(shell string) {
	switch shell {
	case "powershell", "pwsh":
		fmt.Println(`function x {
    param([string]$target)
    $exe = (Get-Command plx.exe -CommandType Application -ErrorAction SilentlyContinue).Source
    if (-not $exe) { return }
    $p = if ($target) { & $exe jump $target } else { & $exe }
    if ($p) {
        $trimmed = ($p | Out-String).Trim()
        if ($trimmed -and (Test-Path -LiteralPath $trimmed)) {
            Set-Location -LiteralPath $trimmed
        }
    }
}
function plx {
    $exe = (Get-Command plx.exe -CommandType Application -ErrorAction SilentlyContinue).Source
    if (-not $exe) { return }
    if ($args.Count -eq 0) {
        $p = & $exe
        if ($p) {
            $trimmed = ($p | Out-String).Trim()
            if ($trimmed -and (Test-Path -LiteralPath $trimmed)) {
                Set-Location -LiteralPath $trimmed
            }
        }
        return
    }
    if ($args[0] -eq 'jump' -and $args.Count -ge 2) {
        $p = & $exe jump $args[1]
        if ($p) {
            $trimmed = ($p | Out-String).Trim()
            if ($trimmed -and (Test-Path -LiteralPath $trimmed)) {
                Set-Location -LiteralPath $trimmed
            }
        }
        return
    }
    & $exe @args
}`)
	case "bash", "zsh":
		fmt.Println(`x() {
    local target
    if [ $# -eq 0 ]; then
        target="$(command plx)"
    else
        target="$(command plx jump "$@")"
    fi
    if [ -n "$target" ] && [ -d "$target" ]; then
        cd "$target" || return 1
    fi
}
plx() {
    if [ $# -eq 0 ]; then
        local target
        target="$(command plx)"
        if [ -n "$target" ] && [ -d "$target" ]; then
            cd "$target" || return 1
        fi
    elif [ "$1" = "jump" ] && [ -n "$2" ]; then
        local target
        target="$(command plx jump "$2")"
        if [ -n "$target" ] && [ -d "$target" ]; then
            cd "$target" || return 1
        fi
    else
        command plx "$@"
    fi
}`)
	case "fish":
		fmt.Println(`function x
    if test (count $argv) -eq 0
        set target (command plx)
    else
        set target (command plx jump $argv)
    end
    if test -n "$target" -a -d "$target"
        cd $target
    end
end
function plx
    if test (count $argv) -eq 0
        set target (command plx)
        if test -n "$target" -a -d "$target"
            cd $target
        end
    else if test "$argv[1]" = "jump" -a (count $argv) -ge 2
        set target (command plx jump $argv[2])
        if test -n "$target" -a -d "$target"
            cd $target
        end
    else
        command plx $argv
    end
end`)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}