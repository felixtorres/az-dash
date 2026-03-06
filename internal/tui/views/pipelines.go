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

type PipelineView struct {
	sections      []pipeSection
	activeSection int
	cursor        int
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
	}
}

func (v *PipelineView) PrevSection() {
	if len(v.sections) > 0 {
		v.activeSection = (v.activeSection - 1 + len(v.sections)) % len(v.sections)
		v.cursor = 0
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

func (v *PipelineView) CursorFirst() { v.cursor = 0 }

func (v *PipelineView) CursorLast() {
	s := v.currentSection()
	if s != nil && len(s.data) > 0 {
		v.cursor = len(s.data) - 1
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
	if v.showPreview {
		tableWidth = v.width * 55 / 100
	}

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
	var rows []string

	nameWidth := width - 60
	header := fmt.Sprintf("  %-*s %-18s %-10s %-10s %-10s",
		nameWidth, "Pipeline", "Branch", "Status", "Duration", "Finished")
	rows = append(rows, v.theme.FaintText.Render(header))

	for i, build := range builds {
		name := utils.TruncateString(build.Definition.Name, nameWidth)
		branch := utils.TruncateString(trimRef(build.SourceBranch), 18)

		status := build.Status
		if build.Result != "" {
			status = build.Result
		}

		duration := utils.FormatDuration(build.Duration())

		finished := "-"
		if build.FinishTime != nil {
			finished = utils.RelativeTime(*build.FinishTime)
		}

		row := fmt.Sprintf("  %-*s %-18s %-10s %-10s %-10s",
			nameWidth, name, branch, status, duration, finished)

		if i == v.cursor {
			rows = append(rows, v.theme.SelectedRow.Render(row))
		} else {
			rows = append(rows, v.buildStatusStyle(build).Render(row))
		}
	}

	return strings.Join(rows, "\n")
}

func (v *PipelineView) renderPreview(width int) string {
	build := v.SelectedBuild()
	if build == nil {
		return ""
	}

	var lines []string
	lines = append(lines, v.theme.Title.Render(utils.TruncateString(build.Definition.Name, width)))
	lines = append(lines, "")
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("Run #%d  %s", build.ID, build.BuildNumber)))
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("Branch: %s", trimRef(build.SourceBranch))))
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("Triggered by: %s", build.RequestedFor.DisplayName)))

	status := build.Status
	if build.Result != "" {
		status = build.Result
	}
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("Status: %s", status)))

	if build.Duration() > 0 {
		lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("Duration: %s", utils.FormatDuration(build.Duration()))))
	}

	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("Commit: %s", utils.TruncateString(build.SourceVersion, 8))))

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
