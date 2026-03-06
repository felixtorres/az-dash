package theme

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/felixtorres/az-dash/internal/config"
)

type Theme struct {
	Text          lipgloss.Style
	FaintText     lipgloss.Style
	Border        lipgloss.Style
	SelectedRow   lipgloss.Style
	Title         lipgloss.Style
	SectionTitle  lipgloss.Style
	ActiveTab     lipgloss.Style
	InactiveTab   lipgloss.Style
	StatusBar     lipgloss.Style
	PROpen        lipgloss.Style
	PRDraft       lipgloss.Style
	PRCompleted   lipgloss.Style
	PRAbandoned   lipgloss.Style
	WINew         lipgloss.Style
	WIActive      lipgloss.Style
	WIResolved    lipgloss.Style
	WIClosed      lipgloss.Style
	PipeSucceeded lipgloss.Style
	PipeFailed    lipgloss.Style
	PipeRunning   lipgloss.Style
	PipeCanceled  lipgloss.Style
	Spinner       lipgloss.Style
	Error         lipgloss.Style
}

var defaultColors = map[string]string{
	"text":      "#e0e0e0",
	"faintText": "#888888",
	"border":    "#444444",
	"selected":  "#58a6ff",
	"title":     "#ffffff",
	"prOpen":    "#3fb950",
	"prDraft":   "#8b949e",
	"prDone":    "#a371f7",
	"prAbort":   "#f85149",
	"wiNew":     "#58a6ff",
	"wiActive":  "#3fb950",
	"wiResolved": "#a371f7",
	"wiClosed":  "#8b949e",
	"pipeOk":    "#3fb950",
	"pipeFail":  "#f85149",
	"pipeRun":   "#d29922",
	"pipeCancel": "#8b949e",
}

func New(cfg config.Theme) *Theme {
	c := func(cfgVal, fallback string) lipgloss.Color {
		if cfgVal != "" {
			return lipgloss.Color(cfgVal)
		}
		return lipgloss.Color(fallback)
	}

	return &Theme{
		Text:          lipgloss.NewStyle().Foreground(c(cfg.Colors.Text, defaultColors["text"])),
		FaintText:     lipgloss.NewStyle().Foreground(c(cfg.Colors.FaintText, defaultColors["faintText"])),
		Border:        lipgloss.NewStyle().BorderForeground(c(cfg.Colors.Border, defaultColors["border"])),
		SelectedRow:   lipgloss.NewStyle().Background(lipgloss.Color("#2d333b")).Bold(true),
		Title:         lipgloss.NewStyle().Foreground(c(cfg.Colors.Title, defaultColors["title"])).Bold(true),
		SectionTitle:  lipgloss.NewStyle().Foreground(c(cfg.Colors.Selected, defaultColors["selected"])).Bold(true).Padding(0, 1),
		ActiveTab:     lipgloss.NewStyle().Foreground(c(cfg.Colors.Selected, defaultColors["selected"])).Bold(true).Underline(true).Padding(0, 2),
		InactiveTab:   lipgloss.NewStyle().Foreground(c(cfg.Colors.FaintText, defaultColors["faintText"])).Padding(0, 2),
		StatusBar:     lipgloss.NewStyle().Foreground(c(cfg.Colors.FaintText, defaultColors["faintText"])),
		PROpen:        lipgloss.NewStyle().Foreground(c(cfg.Colors.PR.Open, defaultColors["prOpen"])),
		PRDraft:       lipgloss.NewStyle().Foreground(c(cfg.Colors.PR.Draft, defaultColors["prDraft"])),
		PRCompleted:   lipgloss.NewStyle().Foreground(c(cfg.Colors.PR.Completed, defaultColors["prDone"])),
		PRAbandoned:   lipgloss.NewStyle().Foreground(c(cfg.Colors.PR.Abandoned, defaultColors["prAbort"])),
		WINew:         lipgloss.NewStyle().Foreground(c(cfg.Colors.WorkItem.New, defaultColors["wiNew"])),
		WIActive:      lipgloss.NewStyle().Foreground(c(cfg.Colors.WorkItem.Active, defaultColors["wiActive"])),
		WIResolved:    lipgloss.NewStyle().Foreground(c(cfg.Colors.WorkItem.Resolved, defaultColors["wiResolved"])),
		WIClosed:      lipgloss.NewStyle().Foreground(c(cfg.Colors.WorkItem.Closed, defaultColors["wiClosed"])),
		PipeSucceeded: lipgloss.NewStyle().Foreground(c(cfg.Colors.Pipeline.Succeeded, defaultColors["pipeOk"])),
		PipeFailed:    lipgloss.NewStyle().Foreground(c(cfg.Colors.Pipeline.Failed, defaultColors["pipeFail"])),
		PipeRunning:   lipgloss.NewStyle().Foreground(c(cfg.Colors.Pipeline.Running, defaultColors["pipeRun"])),
		PipeCanceled:  lipgloss.NewStyle().Foreground(c(cfg.Colors.Pipeline.Canceled, defaultColors["pipeCancel"])),
		Spinner:       lipgloss.NewStyle().Foreground(c(cfg.Colors.Selected, defaultColors["selected"])),
		Error:         lipgloss.NewStyle().Foreground(lipgloss.Color("#f85149")),
	}
}
