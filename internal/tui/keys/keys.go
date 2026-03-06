package keys

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	FirstLine     key.Binding
	LastLine      key.Binding
	PageDown      key.Binding
	PageUp        key.Binding
	NextSection   key.Binding
	PrevSection   key.Binding
	TogglePreview key.Binding
	Refresh       key.Binding
	RefreshAll    key.Binding
	OpenBrowser   key.Binding
	CopyURL       key.Binding
	CopyID        key.Binding
	Search        key.Binding
	ViewPRs       key.Binding
	ViewWorkItems key.Binding
	ViewPipelines key.Binding
	Help          key.Binding
	Quit          key.Binding
}

var Global = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	FirstLine: key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g", "first"),
	),
	LastLine: key.NewBinding(
		key.WithKeys("G", "end"),
		key.WithHelp("G", "last"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "page up"),
	),
	NextSection: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next section"),
	),
	PrevSection: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev section"),
	),
	TogglePreview: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "preview"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	RefreshAll: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "refresh all"),
	),
	OpenBrowser: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in browser"),
	),
	CopyURL: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "copy URL"),
	),
	CopyID: key.NewBinding(
		key.WithKeys("#"),
		key.WithHelp("#", "copy ID"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	ViewPRs: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "PRs"),
	),
	ViewWorkItems: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "Work Items"),
	),
	ViewPipelines: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "Pipelines"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
