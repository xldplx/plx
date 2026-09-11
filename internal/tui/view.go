package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xldplx/plx/internal/git"
)

// View renders the terminal interface.
func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	var sb strings.Builder

	// 1. Header
	dirtyCount := 0
	for _, r := range m.repos {
		if !r.IsClean {
			dirtyCount++
		}
	}

	headerLeft := titleStyle.Render(" plx : Workspace Cockpit ")
	stats := fmt.Sprintf("%d Repos (%d Dirty)", len(m.repos), dirtyCount)
	if m.loading {
		stats += " [Scanning...]"
	}
	headerRight := badgeStyle.Render(stats)

	headerWidth := m.width
	if headerWidth < 80 {
		headerWidth = 80
	}
	gap := headerWidth - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if gap < 1 {
		gap = 1
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, headerLeft, strings.Repeat("─", gap), headerRight))
	sb.WriteString("\n\n")

	// 2. Filter input box
	filterPrefix := lipgloss.NewStyle().Bold(true).Foreground(colorActive).Render(" Filter: ")
	sb.WriteString(filterPrefix + m.textInput.View() + "\n\n")

	// 3. Repository Table List
	if len(m.filtered) == 0 {
		if m.loading {
			sb.WriteString(ageStyle.Render("  Scanning configured workspace roots for git repositories...\n\n"))
		} else {
			sb.WriteString(ageStyle.Render("  No repositories matched your filter or workspace paths.\n\n"))
		}
	} else {
		// Calculate visible window
		visibleRows := 10
		start := 0
		if m.cursor >= visibleRows {
			start = m.cursor - visibleRows + 1
		}
		end := start + visibleRows
		if end > len(m.filtered) {
			end = len(m.filtered)
		}

		for i := start; i < end; i++ {
			repo := m.filtered[i]
			isSelected := i == m.cursor

			cursor := "  "
			if isSelected {
				cursor = "> "
			}

			// Format columns
			nameCol := fmt.Sprintf("%-22s", truncate(repo.Name, 22))
			branchCol := branchStyle.Render(fmt.Sprintf("%-16s", truncate(repo.Branch, 16)))

			var statusCol string
			if repo.IsClean {
				statusCol = cleanBadgeStyle.Render("clean                  ")
			} else {
				statusCol = dirtyBadgeStyle.Render(fmt.Sprintf("* %-21s", truncate(repo.DirtySummary, 21)))
			}

			syncCol := fmt.Sprintf("%s %s",
				aheadStyle.Render(fmt.Sprintf("↑%d", repo.Ahead)),
				behindStyle.Render(fmt.Sprintf("↓%d", repo.Behind)),
			)
			syncCol = fmt.Sprintf("%-10s", syncCol)

			ageCol := ageStyle.Render(fmt.Sprintf("%10s", repo.LastCommitAge))

			rowContent := fmt.Sprintf("%s%s %s %s %s %s", cursor, nameCol, branchCol, statusCol, syncCol, ageCol)

			if isSelected {
				sb.WriteString(selectedRowStyle.Render(rowContent))
			} else {
				sb.WriteString(normalRowStyle.Render(rowContent))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")

	// 4. Details / Preview Pane
	var selectedRepo *git.Repo
	if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		selectedRepo = m.filtered[m.cursor]
	}

	previewContent := renderPreview(selectedRepo, headerWidth-4)
	sb.WriteString(previewBoxStyle.Width(headerWidth - 4).Render(previewContent))
	sb.WriteString("\n")

	// 5. Footer & Keybindings
	footer := fmt.Sprintf(
		" %s Jump  %s Editor  %s Filter  %s Rescan  %s Quit",
		keyStyle.Render("[Enter]"),
		keyStyle.Render("[o]"),
		keyStyle.Render("[/]"),
		keyStyle.Render("[r]"),
		keyStyle.Render("[q]"),
	)
	if m.statusMsg != "" {
		footer += "  " + lipgloss.NewStyle().Foreground(colorYellow).Render("• "+m.statusMsg)
	}
	sb.WriteString(footerStyle.Render(footer) + "\n")

	return sb.String()
}

func renderPreview(r *git.Repo, width int) string {
	if r == nil {
		return "Select a repository to view details."
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Path: %s", r.Path))

	if r.LastCommitMsg != "" {
		lines = append(lines, fmt.Sprintf("Last Commit: %s (%s)", r.LastCommitMsg, r.LastCommitAge))
	}

	if len(r.ChangedFiles) > 0 {
		lines = append(lines, "Changes: "+strings.Join(r.ChangedFiles, "  "))
	} else {
		lines = append(lines, "Working tree is clean.")
	}

	return strings.Join(lines, "\n")
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
