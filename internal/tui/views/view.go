package views

// View is the interface all view types implement for navigation.
type View interface {
	NextSection()
	PrevSection()
	CursorUp()
	CursorDown()
	CursorFirst()
	CursorLast()
	TogglePreview()
	SetSize(width, height int)
	View() string
}
