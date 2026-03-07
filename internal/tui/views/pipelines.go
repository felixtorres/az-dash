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

type PipelineView struct {
	sections      []pipeSection
	activeSection int
	cursor        int
	scrollOffset  int
	showPreview   bool
	width, height int
	theme         *theme.Theme
}

type pipeSection struct {
	title   string
	config  config.PipelineSection
	data    []azdo.Build
	err     error
	loading bool
}

func NewPipelineView(sections []config.PipelineSection, th *theme.Theme) *PipelineView {
	ss := make([]pipeSection, len(sections))
	for i, s := range sections {
		ss[i] = pipeSection{title: s.Title, config: s, loading: true}
	}
	return &PipelineView{
		sections:    ss,
		showPreview: true,
		theme:       th,
	}
}

func (v *PipelineView) SetSectionData(index int, data []azdo.Build) {
	if index < len(v.sections) {
		v.sections[index].data = data
		v.sections[index].loading = false
		v.sections[index].err = nil
	}
}

func (v *PipelineView) SetSectionError(index int, err error) {
	if index < len(v.sections) {
		v.sections[index].err = err
		v.sections[index].loading = false
	}
}

func (v *PipelineView) NextSection() {
	if len(v.sections) > 0 {
		v.activeSection = (v.activeSection + 1) % len(v.sections)
		v.cursor = 0
		v.scrollOffset = 0
	}
}

func (v *PipelineView) PrevSection() {
	if len(v.sections) > 0 {
		v.activeSection = (v.activeSection - 1 + len(v.sections)) % len(v.sections)
		v.cursor = 0
		v.scrollOffset = 0
	}
}

func (v *PipelineView) CursorUp() {
	if v.cursor > 0 {
		v.cursor--
	}
}

func (v *PipelineView) CursorDown() {
	s := v.currentSection()
	if s != nil && v.cursor < len(s.data)-1 {
		v.cursor++
	}
}

func (v *PipelineView) ActiveSectionIndex() int { return v.activeSection }
func (v *PipelineView) CursorFirst()            { v.cursor = 0; v.scrollOffset = 0 }

func (v *PipelineView) CursorLast() {
	s := v.currentSection()
	if s != nil && len(s.data) > 0 {
		v.cursor = len(s.data) - 1
	}
}

func (v *PipelineView) PageDown() {
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

func (v *PipelineView) PageUp() {
	pageSize := v.height - 5
	if pageSize < 1 {
		pageSize = 1
	}
	v.cursor -= pageSize
	if v.cursor < 0 {
		v.cursor = 0
	}
}

func (v *PipelineView) TogglePreview() { v.showPreview = !v.showPreview }
func (v *PipelineView) SetSize(w, h int) { v.width = w; v.height = h }

func (v *PipelineView) currentSection() *pipeSection {
	if v.activeSection < len(v.sections) {
		return &v.sections[v.activeSection]
	}
	return nil
}

func (v *PipelineView) SelectedBuild() *azdo.Build {
	s := v.currentSection()
	if s != nil && v.cursor < len(s.data) {
		return &s.data[v.cursor]
	}
	return nil
}

func (v *PipelineView) View() string {
	if len(v.sections) == 0 {
		return "  No pipeline sections configured."
	}

	var b strings.Builder
	b.WriteString(v.renderSectionTabs())
	b.WriteString("\n")

	s := v.currentSection()
	if s == nil {
		return b.String()
	}

	if s.loading {
		b.WriteString("  Loading pipeline runs...")
		return b.String()
	}

	if s.err != nil {
		b.WriteString(fmt.Sprintf("  %s", v.theme.Error.Render(fmt.Sprintf("Error: %v", s.err))))
		return b.String()
	}

	if len(s.data) == 0 {
		b.WriteString("  No pipeline runs match this filter.")
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

func (v *PipelineView) renderSectionTabs() string {
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

func (v *PipelineView) renderTable(builds []azdo.Build, width int) string {
	columns := []table.Column{
		{Title: "", Width: 2},           // status icon
		{Title: "Pipeline"},             // flex
		{Title: "Branch", Width: 18},
		{Title: "Status", Width: 10},
		{Title: "Duration", Width: 8},
		{Title: "Triggered By", Width: 16},
		{Title: "Finished", Width: 8},
	}

	rows := make([]table.Row, len(builds))
	for i, build := range builds {
		icon := buildStatusIcon(build)

		status := build.Result
		if status == "" {
			status = build.Status
		}

		finished := "—"
		if build.FinishTime != nil {
			finished = utils.RelativeTime(*build.FinishTime)
		}

		rows[i] = table.Row{
			Cells: []string{
				icon,
				build.Definition.Name,
				utils.TruncateString(trimRef(build.SourceBranch), 18),
				status,
				utils.FormatDuration(build.Duration()),
				utils.TruncateString(build.RequestedFor.DisplayName, 16),
				finished,
			},
			Style: v.buildStatusStyle(build),
		}
	}

	styles := table.DefaultStyles()
	result, newOffset := table.Render(columns, rows, width, v.height-3, v.cursor, v.scrollOffset, styles)
	v.scrollOffset = newOffset
	return result
}

func (v *PipelineView) renderPreview(width int) string {
	build := v.SelectedBuild()
	if build == nil {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).Width(width)
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#58a6ff"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	var lines []string

	icon := buildStatusIcon(*build)
	status := build.Result
	if status == "" {
		status = build.Status
	}

	lines = append(lines, titleStyle.Render(build.Definition.Name))
	lines = append(lines, "")
	lines = append(lines, dimStyle.Render(fmt.Sprintf("  %s Run #%d  %s", icon, build.ID, build.BuildNumber)))
	lines = append(lines, "")

	fields := []struct{ label, value string }{
		{"Status", fmt.Sprintf("%s %s", icon, status)},
		{"Branch", trimRef(build.SourceBranch)},
		{"Triggered", build.RequestedFor.DisplayName},
		{"Duration", utils.FormatDuration(build.Duration())},
		{"Commit", utils.TruncateString(build.SourceVersion, 8)},
	}
	for _, f := range fields {
		lines = append(lines, fmt.Sprintf("  %s  %s",
			labelStyle.Width(12).Render(f.label), dimStyle.Render(f.value)))
	}

	return strings.Join(lines, "\n")
}

func (v *PipelineView) buildStatusStyle(build azdo.Build) lipgloss.Style {
	result := build.Result
	if result == "" {
		result = build.Status
	}
	switch result {
	case "succeeded":
		return v.theme.PipeSucceeded
	case "failed":
		return v.theme.PipeFailed
	case "inProgress":
		return v.theme.PipeRunning
	case "canceled", "cancelling":
		return v.theme.PipeCanceled
	default:
		return v.theme.Text
	}
}

func buildStatusIcon(build azdo.Build) string {
	result := build.Result
	if result == "" {
		result = build.Status
	}
	switch result {
	case "succeeded":
		return "✓"
	case "partiallySucceeded":
		return "◐"
	case "failed":
		return "✗"
	case "inProgress":
		return "●"
	case "notStarted":
		return "◌"
	case "canceled", "cancelling":
		return "⊘"
	default:
		return "○"
	}
}
