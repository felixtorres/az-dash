package keys

import "github.com/charmbracelet/bubbles/key"

type WorkItemKeyMap struct {
	Assign      key.Binding
	Unassign    key.Binding
	Comment     key.Binding
	ChangeState key.Binding
}

var WorkItem = WorkItemKeyMap{
	Assign: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "assign"),
	),
	Unassign: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "unassign"),
	),
	Comment: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "comment"),
	),
	ChangeState: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "change state"),
	),
}
