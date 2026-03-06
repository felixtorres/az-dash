package keys

import "github.com/charmbracelet/bubbles/key"

type PipelineKeyMap struct {
	Rerun  key.Binding
	Cancel key.Binding
	Logs   key.Binding
}

var Pipeline = PipelineKeyMap{
	Rerun: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "re-run"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "cancel"),
	),
	Logs: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "view logs"),
	),
}
