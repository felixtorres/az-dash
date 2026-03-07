package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Column defines a table column.
type Column struct {
	Title    string
	Width    int    // fixed width; 0 = flex
	MinWidth int
	Hidden   bool
}

// Row is a styled table row.
type Row struct {
	Cells    []string
	Style    lipgloss.Style
	Selected bool
}

// Styles controls table appearance.
type Styles struct {
	Header      lipgloss.Style
	Cell        lipgloss.Style
	Selected    lipgloss.Style
	Border      lipgloss.Style
	BorderChar  string
	HeaderSep   string
}

func DefaultStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#e0e0e0")).
			Background(lipgloss.Color("#1c2128")).
			Padding(0, 1),
		Cell: lipgloss.NewStyle().
			Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Background(lipgloss.Color("#2d4f67")).
			Bold(true).
			Padding(0, 1),
		Border: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")),
		BorderChar: "─",
		HeaderSep:  "─",
	}
}

// Render builds the table string.
func Render(columns []Column, rows []Row, width, maxHeight int, cursor, offset int, styles Styles) (string, int) {
	if len(columns) == 0 || width == 0 {
		return "", offset
	}

	// Calculate column widths
	colWidths := resolveWidths(columns, width)

	var lines []string

	// Header
	headerLine := renderRow(columns, colWidths, headerCells(columns), styles.Header)
	lines = append(lines, headerLine)

	// Header separator
	sepParts := make([]string, len(columns))
	for i, cw := range colWidths {
		if columns[i].Hidden {
			continue
		}
		sepParts[i] = strings.Repeat(styles.HeaderSep, cw+2) // +2 for padding
	}
	sep := styles.Border.Render(joinVisible(sepParts, columns, "┼"))
	lines = append(lines, sep)

	// Visible rows with viewport scrolling
	visibleRows := maxHeight - 2 // header + separator
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Adjust offset to keep cursor visible
	if cursor < offset {
		offset = cursor
	}
	if cursor >= offset+visibleRows {
		offset = cursor - visibleRows + 1
	}
	if offset < 0 {
		offset = 0
	}

	end := offset + visibleRows
	if end > len(rows) {
		end = len(rows)
	}

	for i := offset; i < end; i++ {
		row := rows[i]
		style := styles.Cell
		if row.Selected || i == cursor {
			style = styles.Selected
		} else if row.Style.Value() != "" {
			style = row.Style.Inherit(styles.Cell)
		}
		lines = append(lines, renderRow(columns, colWidths, row.Cells, style))
	}

	// Scroll indicator
	if len(rows) > visibleRows {
		pos := ""
		if offset > 0 && end < len(rows) {
			pos = "↑↓"
		} else if offset > 0 {
			pos = "↑"
		} else if end < len(rows) {
			pos = "↓"
		}
		if pos != "" {
			indicator := styles.Border.Render(
				strings.Repeat(" ", width-len(pos)-1) + pos)
			lines = append(lines, indicator)
		}
	}

	return strings.Join(lines, "\n"), offset
}

func headerCells(columns []Column) []string {
	cells := make([]string, len(columns))
	for i, c := range columns {
		cells[i] = c.Title
	}
	return cells
}

func renderRow(columns []Column, widths []int, cells []string, style lipgloss.Style) string {
	parts := make([]string, 0, len(columns))
	for i, col := range columns {
		if col.Hidden {
			continue
		}
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		w := widths[i]
		// Truncate or pad
		if len(cell) > w {
			if w > 3 {
				cell = cell[:w-3] + "..."
			} else {
				cell = cell[:w]
			}
		}
		parts = append(parts, style.Width(w+2).MaxWidth(w+2).Render(cell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func resolveWidths(columns []Column, totalWidth int) []int {
	widths := make([]int, len(columns))
	usedWidth := 0
	flexCount := 0

	for i, col := range columns {
		if col.Hidden {
			continue
		}
		if col.Width > 0 {
			widths[i] = col.Width
			usedWidth += col.Width + 2 // padding
		} else {
			flexCount++
		}
	}

	remaining := totalWidth - usedWidth
	if flexCount > 0 && remaining > 0 {
		flexWidth := remaining / flexCount
		for i, col := range columns {
			if col.Hidden || col.Width > 0 {
				continue
			}
			w := flexWidth - 2 // padding
			if col.MinWidth > 0 && w < col.MinWidth {
				w = col.MinWidth
			}
			if w < 4 {
				w = 4
			}
			widths[i] = w
		}
	}

	return widths
}

func joinVisible(parts []string, columns []Column, sep string) string {
	var visible []string
	for i, p := range parts {
		if !columns[i].Hidden {
			visible = append(visible, p)
		}
	}
	return strings.Join(visible, sep)
}
