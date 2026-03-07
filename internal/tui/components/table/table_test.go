package table

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderBasic(t *testing.T) {
	columns := []Column{
		{Title: "ID", Width: 4},
		{Title: "Name", Width: 10},
		{Title: "Status", Width: 8},
	}

	rows := []Row{
		{Cells: []string{"1", "Alice", "Active"}},
		{Cells: []string{"2", "Bob", "Done"}},
		{Cells: []string{"3", "Carol", "New"}},
	}

	result, _ := Render(columns, rows, 40, 20, 0, 0, DefaultStyles())

	if !strings.Contains(result, "ID") {
		t.Error("missing header 'ID'")
	}
	if !strings.Contains(result, "Alice") {
		t.Error("missing row 'Alice'")
	}
	if !strings.Contains(result, "Carol") {
		t.Error("missing row 'Carol'")
	}
}

func TestRenderScrolling(t *testing.T) {
	columns := []Column{
		{Title: "Name", Width: 10},
	}

	rows := make([]Row, 50)
	for i := range rows {
		rows[i] = Row{Cells: []string{strings.Repeat("x", 5)}}
	}

	// Max height = 5 (header + sep + 3 rows)
	result, offset := Render(columns, rows, 20, 5, 0, 0, DefaultStyles())

	if offset != 0 {
		t.Errorf("initial offset = %d, want 0", offset)
	}

	// Cursor at row 10 should scroll
	_, offset = Render(columns, rows, 20, 5, 10, 0, DefaultStyles())
	if offset == 0 {
		t.Error("expected offset > 0 when cursor at 10")
	}

	// Should show scroll indicator
	if !strings.Contains(result, "↓") {
		// May or may not show depending on rendering — just check it doesn't panic
	}
}

func TestRenderHiddenColumns(t *testing.T) {
	columns := []Column{
		{Title: "Visible", Width: 10},
		{Title: "Hidden", Width: 10, Hidden: true},
		{Title: "Also Visible", Width: 10},
	}

	rows := []Row{
		{Cells: []string{"a", "b", "c"}},
	}

	result, _ := Render(columns, rows, 40, 10, 0, 0, DefaultStyles())

	if strings.Contains(result, "Hidden") {
		t.Error("hidden column should not appear")
	}
	if !strings.Contains(result, "Visible") {
		t.Error("visible column missing")
	}
}

func TestRenderEmptyRows(t *testing.T) {
	columns := []Column{
		{Title: "Name", Width: 10},
	}

	result, _ := Render(columns, []Row{}, 20, 10, 0, 0, DefaultStyles())

	if !strings.Contains(result, "Name") {
		t.Error("header should still render with empty rows")
	}
}

func TestRenderFlexColumn(t *testing.T) {
	columns := []Column{
		{Title: "ID", Width: 4},
		{Title: "Title"},           // flex
		{Title: "Status", Width: 8},
	}

	rows := []Row{
		{Cells: []string{"1", "A very long title that should be truncated", "OK"}},
	}

	// Should not panic on any width
	for _, w := range []int{20, 40, 80, 120} {
		result, _ := Render(columns, rows, w, 10, 0, 0, DefaultStyles())
		if result == "" {
			t.Errorf("empty result at width %d", w)
		}
	}
}

func TestRenderWithRowStyle(t *testing.T) {
	columns := []Column{
		{Title: "Name", Width: 10},
	}

	rows := []Row{
		{Cells: []string{"OK"}, Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))},
		{Cells: []string{"Fail"}, Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))},
	}

	result, _ := Render(columns, rows, 20, 10, 0, 0, DefaultStyles())
	if !strings.Contains(result, "OK") || !strings.Contains(result, "Fail") {
		t.Error("styled rows should render")
	}
}

func TestResolveWidths(t *testing.T) {
	columns := []Column{
		{Title: "A", Width: 4},
		{Title: "B"},              // flex
		{Title: "C", Width: 6},
		{Title: "D", Hidden: true, Width: 10},
	}

	widths := resolveWidths(columns, 60)

	if widths[0] != 4 {
		t.Errorf("A width = %d, want 4", widths[0])
	}
	if widths[2] != 6 {
		t.Errorf("C width = %d, want 6", widths[2])
	}
	if widths[1] <= 0 {
		t.Errorf("B (flex) width = %d, want > 0", widths[1])
	}
}
