package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/felixtorres/az-dash/internal/azdo"
	"github.com/felixtorres/az-dash/internal/config"
	"github.com/felixtorres/az-dash/internal/tui/theme"
	"github.com/felixtorres/az-dash/internal/utils"
)

type PRView struct {
	sections       []prSection
	activeSection  int
	cursor         int
	showPreview    bool
	width, height  int
	theme          *theme.Theme
}

type prSection struct {
	title   string
	config  config.PRSection
	data    []azdo.PullRequest
	err     error
	loading bool
}

func NewPRView(sections []config.PRSection, th *theme.Theme) *PRView {
	ss := make([]prSection, len(sections))
	for i, s := range sections {
		ss[i] = prSection{title: s.Title, config: s, loading: true}
	}
	return &PRView{
		sections:    ss,
		showPreview: true,
		theme:       th,
	}
}

func (v *PRView) SetSectionData(index int, data []azdo.PullRequest) {
	if index < len(v.sections) {
		v.sections[index].data = data
		v.sections[index].loading = false
		v.sections[index].err = nil
	}
}

func (v *PRView) SetSectionError(index int, err error) {
	if index < len(v.sections) {
		v.sections[index].err = err
		v.sections[index].loading = false
	}
}

func (v *PRView) NextSection() {
	if len(v.sections) > 0 {
		v.activeSection = (v.activeSection + 1) % len(v.sections)
		v.cursor = 0
	}
}

func (v *PRView) PrevSection() {
	if len(v.sections) > 0 {
		v.activeSection = (v.activeSection - 1 + len(v.sections)) % len(v.sections)
		v.cursor = 0
	}
}

func (v *PRView) CursorUp() {
	if v.cursor > 0 {
		v.cursor--
	}
}

func (v *PRView) CursorDown() {
	s := v.currentSection()
	if s != nil && v.cursor < len(s.data)-1 {
		v.cursor++
	}
}

func (v *PRView) CursorFirst() { v.cursor = 0 }

func (v *PRView) CursorLast() {
	s := v.currentSection()
	if s != nil && len(s.data) > 0 {
		v.cursor = len(s.data) - 1
	}
}

func (v *PRView) TogglePreview() { v.showPreview = !v.showPreview }
func (v *PRView) SetSize(w, h int) { v.width = w; v.height = h }

func (v *PRView) currentSection() *prSection {
	if v.activeSection < len(v.sections) {
		return &v.sections[v.activeSection]
	}
	return nil
}

func (v *PRView) SelectedPR() *azdo.PullRequest {
	s := v.currentSection()
	if s != nil && v.cursor < len(s.data) {
		return &s.data[v.cursor]
	}
	return nil
}

func (v *PRView) View() string {
	if len(v.sections) == 0 {
		return "  No PR sections configured."
	}

	var b strings.Builder

	// Section tabs
	b.WriteString(v.renderSectionTabs())
	b.WriteString("\n")

	s := v.currentSection()
	if s == nil {
		return b.String()
	}

	if s.loading {
		b.WriteString("  Loading PRs...")
		return b.String()
	}

	if s.err != nil {
		b.WriteString(fmt.Sprintf("  %s", v.theme.Error.Render(fmt.Sprintf("Error: %v", s.err))))
		return b.String()
	}

	if len(s.data) == 0 {
		b.WriteString("  No pull requests match this filter.")
		return b.String()
	}

	tableWidth := v.width
	if v.showPreview {
		tableWidth = v.width * 55 / 100
	}

	// Table
	table := v.renderTable(s.data, tableWidth)

	if v.showPreview {
		previewWidth := v.width - tableWidth - 3
		preview := v.renderPreview(previewWidth)
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, table, " │ ", preview))
	} else {
		b.WriteString(table)
	}

	return b.String()
}

func (v *PRView) renderSectionTabs() string {
	var tabs []string
	for i, s := range v.sections {
		label := fmt.Sprintf(" %s (%d) ", s.title, len(s.data))
		if i == v.activeSection {
			tabs = append(tabs, v.theme.SectionTitle.Render(label))
		} else {
			tabs = append(tabs, v.theme.FaintText.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (v *PRView) renderTable(prs []azdo.PullRequest, width int) string {
	var rows []string

	// Header
	header := fmt.Sprintf("  %-4s %-*s %-12s %-10s %-10s",
		"#", width-52, "Title", "Author", "Status", "Updated")
	rows = append(rows, v.theme.FaintText.Render(header))

	for i, pr := range prs {
		titleWidth := width - 52
		title := utils.TruncateString(pr.Title, titleWidth)

		status := pr.Status
		if pr.IsDraft {
			status = "draft"
		}

		updated := utils.RelativeTime(pr.CreationDate)
		if pr.ClosedDate != nil {
			updated = utils.RelativeTime(*pr.ClosedDate)
		}

		row := fmt.Sprintf("  %-4d %-*s %-12s %-10s %-10s",
			pr.PullRequestID,
			titleWidth, title,
			utils.TruncateString(pr.CreatedBy.DisplayName, 12),
			status,
			updated,
		)

		if i == v.cursor {
			rows = append(rows, v.theme.SelectedRow.Render(row))
		} else {
			statusStyle := v.prStatusStyle(pr)
			rows = append(rows, statusStyle.Render(row))
		}
	}

	return strings.Join(rows, "\n")
}

func (v *PRView) renderPreview(width int) string {
	pr := v.SelectedPR()
	if pr == nil {
		return ""
	}

	var lines []string
	lines = append(lines, v.theme.Title.Render(utils.TruncateString(pr.Title, width)))
	lines = append(lines, "")
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("#%d by %s", pr.PullRequestID, pr.CreatedBy.DisplayName)))
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("%s → %s", trimRef(pr.SourceRefName), trimRef(pr.TargetRefName))))
	lines = append(lines, "")

	// Reviewers
	if len(pr.Reviewers) > 0 {
		lines = append(lines, v.theme.Title.Render("Reviewers"))
		for _, r := range pr.Reviewers {
			vote := voteString(r.Vote)
			lines = append(lines, fmt.Sprintf("  %s %s", vote, r.DisplayName))
		}
		lines = append(lines, "")
	}

	// Description
	if pr.Description != "" {
		lines = append(lines, v.theme.Title.Render("Description"))
		desc := pr.Description
		if len(desc) > 500 {
			desc = desc[:500] + "..."
		}
		lines = append(lines, desc)
	}

	return strings.Join(lines, "\n")
}

func (v *PRView) prStatusStyle(pr azdo.PullRequest) lipgloss.Style {
	if pr.IsDraft {
		return v.theme.PRDraft
	}
	switch pr.Status {
	case "active":
		return v.theme.PROpen
	case "completed":
		return v.theme.PRCompleted
	case "abandoned":
		return v.theme.PRAbandoned
	default:
		return v.theme.Text
	}
}

func trimRef(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

func voteString(vote int) string {
	switch {
	case vote == 10:
		return "✓"
	case vote == 5:
		return "~"
	case vote == -5:
		return "…"
	case vote == -10:
		return "✗"
	default:
		return "○"
	}
}
