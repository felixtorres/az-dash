package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/felixtorres/az-dash/internal/azdo"
	"github.com/felixtorres/az-dash/internal/config"
	"github.com/felixtorres/az-dash/internal/tui/components/table"
	"github.com/felixtorres/az-dash/internal/tui/theme"
	"github.com/felixtorres/az-dash/internal/utils"
)

type PRView struct {
	sections      []prSection
	activeSection int
	cursor        int
	scrollOffset  int
	showPreview   bool
	width, height int
	theme         *theme.Theme
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
		v.scrollOffset = 0
	}
}

func (v *PRView) PrevSection() {
	if len(v.sections) > 0 {
		v.activeSection = (v.activeSection - 1 + len(v.sections)) % len(v.sections)
		v.cursor = 0
		v.scrollOffset = 0
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

func (v *PRView) ActiveSectionIndex() int { return v.activeSection }
func (v *PRView) CursorFirst()            { v.cursor = 0; v.scrollOffset = 0 }

func (v *PRView) CursorLast() {
	s := v.currentSection()
	if s != nil && len(s.data) > 0 {
		v.cursor = len(s.data) - 1
	}
}

func (v *PRView) PageDown() {
	s := v.currentSection()
	if s == nil {
		return
	}
	pageSize := v.height - 5
	if pageSize < 1 {
		pageSize = 1
	}
	v.cursor += pageSize
	if v.cursor >= len(s.data) {
		v.cursor = len(s.data) - 1
	}
}

func (v *PRView) PageUp() {
	pageSize := v.height - 5
	if pageSize < 1 {
		pageSize = 1
	}
	v.cursor -= pageSize
	if v.cursor < 0 {
		v.cursor = 0
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
	previewWidth := 0
	if v.showPreview {
		tableWidth = v.width * 55 / 100
		previewWidth = v.width - tableWidth - 1
	}

	tableStr := v.renderTable(s.data, tableWidth)

	if v.showPreview && previewWidth > 10 {
		preview := v.renderPreview(previewWidth)
		divider := v.theme.Border.
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderTop(false).
			BorderBottom(false).
			BorderRight(false).
			Height(lipgloss.Height(tableStr)).
			Render("")
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tableStr, divider, preview))
	} else {
		b.WriteString(tableStr)
	}

	return b.String()
}

func (v *PRView) renderSectionTabs() string {
	var tabs []string
	for i, s := range v.sections {
		count := len(s.data)
		label := fmt.Sprintf(" %s (%d) ", s.title, count)
		if i == v.activeSection {
			tabs = append(tabs, v.theme.SectionTitle.Render(label))
		} else {
			tabs = append(tabs, v.theme.FaintText.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (v *PRView) renderTable(prs []azdo.PullRequest, width int) string {
	columns := []table.Column{
		{Title: "", Width: 2},        // status icon
		{Title: "#", Width: 6},       // PR number
		{Title: "Title"},             // flex
		{Title: "Author", Width: 16},
		{Title: "Reviews", Width: 10},
		{Title: "Branch", Width: 16},
		{Title: "Updated", Width: 8},
	}

	rows := make([]table.Row, len(prs))
	for i, pr := range prs {
		icon := prStatusIcon(pr)
		reviews := reviewSummary(pr.Reviewers)
		branch := utils.TruncateString(trimRef(pr.TargetRefName), 16)

		updated := utils.RelativeTime(pr.CreationDate)
		if pr.ClosedDate != nil {
			updated = utils.RelativeTime(*pr.ClosedDate)
		}

		rows[i] = table.Row{
			Cells: []string{
				icon,
				fmt.Sprintf("%d", pr.PullRequestID),
				pr.Title,
				utils.TruncateString(pr.CreatedBy.DisplayName, 16),
				reviews,
				branch,
				updated,
			},
			Style: v.prStatusStyle(pr),
		}
	}

	styles := table.DefaultStyles()
	result, newOffset := table.Render(columns, rows, width, v.height-3, v.cursor, v.scrollOffset, styles)
	v.scrollOffset = newOffset
	return result
}

func (v *PRView) renderPreview(width int) string {
	pr := v.SelectedPR()
	if pr == nil {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).Width(width)
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#58a6ff"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	var lines []string

	// Title
	lines = append(lines, titleStyle.Render(utils.TruncateString(pr.Title, width)))
	lines = append(lines, "")

	// Meta
	statusIcon := prStatusIcon(*pr)
	lines = append(lines, dimStyle.Render(fmt.Sprintf("  %s #%d  %s → %s",
		statusIcon, pr.PullRequestID,
		trimRef(pr.SourceRefName), trimRef(pr.TargetRefName))))
	lines = append(lines, dimStyle.Render(fmt.Sprintf("  by %s  •  %s",
		pr.CreatedBy.DisplayName, utils.RelativeTime(pr.CreationDate))))
	lines = append(lines, "")

	// Reviewers
	if len(pr.Reviewers) > 0 {
		lines = append(lines, labelStyle.Render("  Reviewers"))
		for _, r := range pr.Reviewers {
			vote := voteIcon(r.Vote)
			required := ""
			if r.IsRequired {
				required = " (required)"
			}
			lines = append(lines, fmt.Sprintf("  %s %s%s", vote, r.DisplayName, required))
		}
		lines = append(lines, "")
	}

	// Merge status
	if pr.MergeStatus != "" {
		lines = append(lines, labelStyle.Render("  Merge Status"))
		mergeIcon := "⊘"
		switch pr.MergeStatus {
		case "succeeded":
			mergeIcon = "✓"
		case "conflicts":
			mergeIcon = "✗"
		case "queued":
			mergeIcon = "◌"
		}
		lines = append(lines, fmt.Sprintf("  %s %s", mergeIcon, pr.MergeStatus))
		lines = append(lines, "")
	}

	// Description
	if pr.Description != "" {
		lines = append(lines, labelStyle.Render("  Description"))
		desc := pr.Description
		if len(desc) > 800 {
			desc = desc[:800] + "..."
		}
		// Word-wrap description to preview width
		for _, line := range strings.Split(desc, "\n") {
			if len(line) > width-4 {
				line = line[:width-4]
			}
			lines = append(lines, "  "+line)
		}
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

func prStatusIcon(pr azdo.PullRequest) string {
	if pr.IsDraft {
		return "◇"
	}
	switch pr.Status {
	case "active":
		return "◉"
	case "completed":
		return "✓"
	case "abandoned":
		return "✗"
	default:
		return "○"
	}
}

func reviewSummary(reviewers []azdo.Reviewer) string {
	if len(reviewers) == 0 {
		return "—"
	}
	approved, waiting, rejected := 0, 0, 0
	for _, r := range reviewers {
		switch {
		case r.Vote == 10 || r.Vote == 5:
			approved++
		case r.Vote == -10:
			rejected++
		case r.Vote == -5:
			waiting++
		}
	}
	parts := []string{}
	if approved > 0 {
		parts = append(parts, fmt.Sprintf("✓%d", approved))
	}
	if rejected > 0 {
		parts = append(parts, fmt.Sprintf("✗%d", rejected))
	}
	if waiting > 0 {
		parts = append(parts, fmt.Sprintf("…%d", waiting))
	}
	noResponse := len(reviewers) - approved - rejected - waiting
	if noResponse > 0 {
		parts = append(parts, fmt.Sprintf("○%d", noResponse))
	}
	return strings.Join(parts, " ")
}

func voteIcon(vote int) string {
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

func trimRef(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

func voteString(vote int) string {
	return voteIcon(vote)
}
