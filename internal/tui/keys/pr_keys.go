package keys

import "github.com/charmbracelet/bubbles/key"

type PRKeyMap struct {
	Approve  key.Binding
	Comment  key.Binding
	Diff     key.Binding
	Assign   key.Binding
	Unassign key.Binding
	Complete key.Binding
	Abandon  key.Binding
	Reopen   key.Binding
}

var PR = PRKeyMap{
	Approve: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "approve"),
	),
	Comment: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "comment"),
	),
	Diff: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "diff"),
	),
	Assign: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "assign reviewer"),
	),
	Unassign: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "remove reviewer"),
	),
	Complete: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "complete/merge"),
	),
	Abandon: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "abandon"),
	),
	Reopen: key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "reactivate"),
	),
}
