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

type WorkItemView struct {
	sections      []wiSection
	activeSection int
	cursor        int
	showPreview   bool
	width, height int
	theme         *theme.Theme
}

type wiSection struct {
	title   string
	config  config.WorkItemSection
	data    []azdo.WorkItem
	err     error
	loading bool
}

func NewWorkItemView(sections []config.WorkItemSection, th *theme.Theme) *WorkItemView {
	ss := make([]wiSection, len(sections))
	for i, s := range sections {
		ss[i] = wiSection{title: s.Title, config: s, loading: true}
	}
	return &WorkItemView{
		sections:    ss,
		showPreview: true,
		theme:       th,
	}
}

func (v *WorkItemView) SetSectionData(index int, data []azdo.WorkItem) {
	if index < len(v.sections) {
		v.sections[index].data = data
		v.sections[index].loading = false
		v.sections[index].err = nil
	}
}

func (v *WorkItemView) SetSectionError(index int, err error) {
	if index < len(v.sections) {
		v.sections[index].err = err
		v.sections[index].loading = false
	}
}

func (v *WorkItemView) NextSection() {
	if len(v.sections) > 0 {
		v.activeSection = (v.activeSection + 1) % len(v.sections)
		v.cursor = 0
	}
}

func (v *WorkItemView) PrevSection() {
	if len(v.sections) > 0 {
		v.activeSection = (v.activeSection - 1 + len(v.sections)) % len(v.sections)
		v.cursor = 0
	}
}

func (v *WorkItemView) CursorUp() {
	if v.cursor > 0 {
		v.cursor--
	}
}

func (v *WorkItemView) CursorDown() {
	s := v.currentSection()
	if s != nil && v.cursor < len(s.data)-1 {
		v.cursor++
	}
}

func (v *WorkItemView) CursorFirst() { v.cursor = 0 }

func (v *WorkItemView) CursorLast() {
	s := v.currentSection()
	if s != nil && len(s.data) > 0 {
		v.cursor = len(s.data) - 1
	}
}

func (v *WorkItemView) TogglePreview() { v.showPreview = !v.showPreview }
func (v *WorkItemView) SetSize(w, h int) { v.width = w; v.height = h }

func (v *WorkItemView) currentSection() *wiSection {
	if v.activeSection < len(v.sections) {
		return &v.sections[v.activeSection]
	}
	return nil
}

func (v *WorkItemView) SelectedWorkItem() *azdo.WorkItem {
	s := v.currentSection()
	if s != nil && v.cursor < len(s.data) {
		return &s.data[v.cursor]
	}
	return nil
}

func (v *WorkItemView) View() string {
	if len(v.sections) == 0 {
		return "  No work item sections configured."
	}

	var b strings.Builder
	b.WriteString(v.renderSectionTabs())
	b.WriteString("\n")

	s := v.currentSection()
	if s == nil {
		return b.String()
	}

	if s.loading {
		b.WriteString("  Loading work items...")
		return b.String()
	}

	if s.err != nil {
		b.WriteString(fmt.Sprintf("  %s", v.theme.Error.Render(fmt.Sprintf("Error: %v", s.err))))
		return b.String()
	}

	if len(s.data) == 0 {
		b.WriteString("  No work items match this query.")
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

func (v *WorkItemView) renderSectionTabs() string {
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

func (v *WorkItemView) renderTable(items []azdo.WorkItem, width int) string {
	var rows []string

	titleWidth := width - 48
	header := fmt.Sprintf("  %-6s %-8s %-*s %-10s %-12s",
		"ID", "Type", titleWidth, "Title", "State", "Assigned To")
	rows = append(rows, v.theme.FaintText.Render(header))

	for i, wi := range items {
		wiType := utils.TruncateString(wi.StringField("System.WorkItemType"), 8)
		title := utils.TruncateString(wi.StringField("System.Title"), titleWidth)
		state := utils.TruncateString(wi.StringField("System.State"), 10)
		assignedTo := utils.TruncateString(wi.StringField("System.AssignedTo"), 12)

		row := fmt.Sprintf("  %-6d %-8s %-*s %-10s %-12s",
			wi.ID, wiType, titleWidth, title, state, assignedTo)

		if i == v.cursor {
			rows = append(rows, v.theme.SelectedRow.Render(row))
		} else {
			rows = append(rows, v.wiStateStyle(state).Render(row))
		}
	}

	return strings.Join(rows, "\n")
}

func (v *WorkItemView) renderPreview(width int) string {
	wi := v.SelectedWorkItem()
	if wi == nil {
		return ""
	}

	var lines []string
	lines = append(lines, v.theme.Title.Render(utils.TruncateString(wi.StringField("System.Title"), width)))
	lines = append(lines, "")
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("#%d  %s  %s",
		wi.ID, wi.StringField("System.WorkItemType"), wi.StringField("System.State"))))
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("Assigned: %s", wi.StringField("System.AssignedTo"))))
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("Area: %s", wi.StringField("System.AreaPath"))))
	lines = append(lines, v.theme.FaintText.Render(fmt.Sprintf("Iteration: %s", wi.StringField("System.IterationPath"))))
	lines = append(lines, "")

	desc := wi.StringField("System.Description")
	if desc != "" {
		lines = append(lines, v.theme.Title.Render("Description"))
		if len(desc) > 500 {
			desc = desc[:500] + "..."
		}
		lines = append(lines, desc)
	}

	return strings.Join(lines, "\n")
}

func (v *WorkItemView) wiStateStyle(state string) lipgloss.Style {
	switch strings.ToLower(state) {
	case "new":
		return v.theme.WINew
	case "active":
		return v.theme.WIActive
	case "resolved":
		return v.theme.WIResolved
	case "closed", "done":
		return v.theme.WIClosed
	default:
		return v.theme.Text
	}
}
