package views

// View is the interface all view types implement for navigation.
type View interface {
	NextSection()
	PrevSection()
	ActiveSectionIndex() int
	CursorUp()
	CursorDown()
	CursorFirst()
	CursorLast()
	PageDown()
	PageUp()
	TogglePreview()
	SetSize(width, height int)
	View() string
}
