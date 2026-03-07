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

type WorkItemView struct {
	sections      []wiSection
	activeSection int
	cursor        int
	scrollOffset  int
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
		v.scrollOffset = 0
	}
}

func (v *WorkItemView) PrevSection() {
	if len(v.sections) > 0 {
		v.activeSection = (v.activeSection - 1 + len(v.sections)) % len(v.sections)
		v.cursor = 0
		v.scrollOffset = 0
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

func (v *WorkItemView) ActiveSectionIndex() int { return v.activeSection }
func (v *WorkItemView) CursorFirst()            { v.cursor = 0; v.scrollOffset = 0 }

func (v *WorkItemView) CursorLast() {
	s := v.currentSection()
	if s != nil && len(s.data) > 0 {
		v.cursor = len(s.data) - 1
	}
}

func (v *WorkItemView) PageDown() {
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

func (v *WorkItemView) PageUp() {
	pageSize := v.height - 5
	if pageSize < 1 {
		pageSize = 1
	}
	v.cursor -= pageSize
	if v.cursor < 0 {
		v.cursor = 0
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
	columns := []table.Column{
		{Title: "", Width: 2},          // type icon
		{Title: "ID", Width: 7},
		{Title: "Type", Width: 8},
		{Title: "Title"},               // flex
		{Title: "State", Width: 10},
		{Title: "Assigned To", Width: 16},
		{Title: "Priority", Width: 4},
	}

	rows := make([]table.Row, len(items))
	for i, wi := range items {
		wiType := wi.StringField("System.WorkItemType")
		icon := wiTypeIcon(wiType)
		state := wi.StringField("System.State")
		priority := wi.StringField("Microsoft.VSTS.Common.Priority")

		rows[i] = table.Row{
			Cells: []string{
				icon,
				fmt.Sprintf("%d", wi.ID),
				utils.TruncateString(wiType, 8),
				wi.StringField("System.Title"),
				state,
				utils.TruncateString(wi.StringField("System.AssignedTo"), 16),
				priority,
			},
			Style: v.wiStateStyle(state),
		}
	}

	styles := table.DefaultStyles()
	result, newOffset := table.Render(columns, rows, width, v.height-3, v.cursor, v.scrollOffset, styles)
	v.scrollOffset = newOffset
	return result
}

func (v *WorkItemView) renderPreview(width int) string {
	wi := v.SelectedWorkItem()
	if wi == nil {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).Width(width)
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#58a6ff"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	var lines []string

	wiType := wi.StringField("System.WorkItemType")
	state := wi.StringField("System.State")

	lines = append(lines, titleStyle.Render(utils.TruncateString(wi.StringField("System.Title"), width)))
	lines = append(lines, "")
	lines = append(lines, dimStyle.Render(fmt.Sprintf("  %s #%d  %s  [%s]",
		wiTypeIcon(wiType), wi.ID, wiType, state)))
	lines = append(lines, "")

	// Fields
	fields := []struct{ label, field string }{
		{"Assigned To", "System.AssignedTo"},
		{"Area", "System.AreaPath"},
		{"Iteration", "System.IterationPath"},
		{"Priority", "Microsoft.VSTS.Common.Priority"},
		{"Tags", "System.Tags"},
	}
	for _, f := range fields {
		val := wi.StringField(f.field)
		if val != "" {
			lines = append(lines, fmt.Sprintf("  %s  %s",
				labelStyle.Width(14).Render(f.label), dimStyle.Render(val)))
		}
	}
	lines = append(lines, "")

	// Description
	desc := wi.StringField("System.Description")
	if desc != "" {
		lines = append(lines, labelStyle.Render("  Description"))
		// Strip HTML tags (basic)
		desc = stripHTML(desc)
		if len(desc) > 800 {
			desc = desc[:800] + "..."
		}
		for _, line := range strings.Split(desc, "\n") {
			if len(line) > width-4 {
				line = line[:width-4]
			}
			lines = append(lines, "  "+line)
		}
	}

	return strings.Join(lines, "\n")
}

func (v *WorkItemView) wiStateStyle(state string) lipgloss.Style {
	switch strings.ToLower(state) {
	case "new", "to do":
		return v.theme.WINew
	case "active", "doing", "in progress":
		return v.theme.WIActive
	case "resolved", "done":
		return v.theme.WIResolved
	case "closed", "removed":
		return v.theme.WIClosed
	default:
		return v.theme.Text
	}
}

func wiTypeIcon(wiType string) string {
	switch strings.ToLower(wiType) {
	case "bug":
		return "🐛"
	case "task":
		return "☑"
	case "user story", "story":
		return "📖"
	case "epic":
		return "⚡"
	case "feature":
		return "★"
	case "issue":
		return "⚠"
	default:
		return "•"
	}
}

func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}
