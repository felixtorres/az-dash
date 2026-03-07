package views

import (
	"strings"
	"testing"

	"github.com/felixtorres/az-dash/internal/azdo"
	"github.com/felixtorres/az-dash/internal/config"
)

func newTestWIView() *WorkItemView {
	return NewWorkItemView([]config.WorkItemSection{
		{Title: "My Tasks", WIQL: "SELECT 1"},
		{Title: "Bugs", WIQL: "SELECT 2"},
	}, newTestTheme())
}

func sampleWorkItems() []azdo.WorkItem {
	return []azdo.WorkItem{
		{ID: 100, Fields: map[string]interface{}{
			"System.Title":        "Fix login",
			"System.WorkItemType": "Bug",
			"System.State":        "Active",
			"System.AssignedTo":   map[string]interface{}{"displayName": "Alice"},
		}},
		{ID: 200, Fields: map[string]interface{}{
			"System.Title":        "Add export",
			"System.WorkItemType": "Task",
			"System.State":        "New",
			"System.AssignedTo":   map[string]interface{}{"displayName": "Bob"},
		}},
	}
}

func TestWIViewSectionNavigation(t *testing.T) {
	v := newTestWIView()

	if v.ActiveSectionIndex() != 0 {
		t.Errorf("initial = %d", v.ActiveSectionIndex())
	}

	v.NextSection()
	if v.ActiveSectionIndex() != 1 {
		t.Errorf("after Next = %d", v.ActiveSectionIndex())
	}

	v.NextSection() // wrap
	if v.ActiveSectionIndex() != 0 {
		t.Errorf("after wrap = %d", v.ActiveSectionIndex())
	}
}

func TestWIViewCursorNavigation(t *testing.T) {
	v := newTestWIView()
	v.SetSize(120, 40)
	v.SetSectionData(0, sampleWorkItems())

	wi := v.SelectedWorkItem()
	if wi.ID != 100 {
		t.Errorf("initial = %d", wi.ID)
	}

	v.CursorDown()
	wi = v.SelectedWorkItem()
	if wi.ID != 200 {
		t.Errorf("after down = %d", wi.ID)
	}

	v.CursorDown() // at end
	wi = v.SelectedWorkItem()
	if wi.ID != 200 {
		t.Errorf("at end = %d", wi.ID)
	}

	v.CursorFirst()
	wi = v.SelectedWorkItem()
	if wi.ID != 100 {
		t.Errorf("after first = %d", wi.ID)
	}
}

func TestWIViewLoadingState(t *testing.T) {
	v := newTestWIView()
	v.SetSize(120, 40)

	output := v.View()
	if !strings.Contains(output, "Loading work items") {
		t.Errorf("expected loading: %q", output)
	}
}

func TestWIViewEmptyState(t *testing.T) {
	v := newTestWIView()
	v.SetSize(120, 40)
	v.SetSectionData(0, []azdo.WorkItem{})

	output := v.View()
	if !strings.Contains(output, "No work items") {
		t.Errorf("expected empty: %q", output)
	}
}

func TestWIViewRendersTable(t *testing.T) {
	v := newTestWIView()
	v.SetSize(120, 40)
	v.SetSectionData(0, sampleWorkItems())

	output := v.View()
	if !strings.Contains(output, "Fix login") {
		t.Errorf("missing title in output")
	}
	if !strings.Contains(output, "100") {
		t.Errorf("missing ID in output")
	}
}

func TestWIViewSelectedNilWhenEmpty(t *testing.T) {
	v := newTestWIView()
	v.SetSectionData(0, []azdo.WorkItem{})

	if v.SelectedWorkItem() != nil {
		t.Error("should be nil")
	}
}
