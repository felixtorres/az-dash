package views

import (
	"strings"
	"testing"
	"time"

	"github.com/felixtorres/az-dash/internal/azdo"
	"github.com/felixtorres/az-dash/internal/config"
)

func newTestPipeView() *PipelineView {
	return NewPipelineView([]config.PipelineSection{
		{Title: "My Runs"},
		{Title: "Failed"},
	}, newTestTheme())
}

func sampleBuilds() []azdo.Build {
	start := time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC)
	finish := time.Date(2026, 3, 6, 10, 5, 0, 0, time.UTC)
	return []azdo.Build{
		{
			ID: 1, BuildNumber: "20260306.1",
			Status: "completed", Result: "succeeded",
			SourceBranch: "refs/heads/main",
			StartTime: &start, FinishTime: &finish,
			Definition:   azdo.BuildDefinition{ID: 10, Name: "CI"},
			RequestedFor: azdo.IdentityRef{DisplayName: "Alice"},
		},
		{
			ID: 2, BuildNumber: "20260306.2",
			Status: "completed", Result: "failed",
			SourceBranch: "refs/heads/feature",
			StartTime: &start, FinishTime: &finish,
			Definition:   azdo.BuildDefinition{ID: 10, Name: "CI"},
			RequestedFor: azdo.IdentityRef{DisplayName: "Bob"},
		},
	}
}

func TestPipeViewSectionNavigation(t *testing.T) {
	v := newTestPipeView()

	v.NextSection()
	if v.ActiveSectionIndex() != 1 {
		t.Errorf("after Next = %d", v.ActiveSectionIndex())
	}
	v.NextSection() // wrap
	if v.ActiveSectionIndex() != 0 {
		t.Errorf("after wrap = %d", v.ActiveSectionIndex())
	}
}

func TestPipeViewCursorNavigation(t *testing.T) {
	v := newTestPipeView()
	v.SetSize(120, 40)
	v.SetSectionData(0, sampleBuilds())

	b := v.SelectedBuild()
	if b.ID != 1 {
		t.Errorf("initial = %d", b.ID)
	}

	v.CursorDown()
	b = v.SelectedBuild()
	if b.ID != 2 {
		t.Errorf("after down = %d", b.ID)
	}

	v.CursorLast()
	b = v.SelectedBuild()
	if b.ID != 2 {
		t.Errorf("after last = %d", b.ID)
	}

	v.CursorFirst()
	b = v.SelectedBuild()
	if b.ID != 1 {
		t.Errorf("after first = %d", b.ID)
	}
}

func TestPipeViewLoadingState(t *testing.T) {
	v := newTestPipeView()
	v.SetSize(120, 40)

	output := v.View()
	if !strings.Contains(output, "Loading pipeline") {
		t.Errorf("expected loading: %q", output)
	}
}

func TestPipeViewEmptyState(t *testing.T) {
	v := newTestPipeView()
	v.SetSize(120, 40)
	v.SetSectionData(0, []azdo.Build{})

	output := v.View()
	if !strings.Contains(output, "No pipeline runs") {
		t.Errorf("expected empty: %q", output)
	}
}

func TestPipeViewRendersTable(t *testing.T) {
	v := newTestPipeView()
	v.SetSize(120, 40)
	v.SetSectionData(0, sampleBuilds())

	output := v.View()
	if !strings.Contains(output, "CI") {
		t.Errorf("missing pipeline name")
	}
	if !strings.Contains(output, "main") {
		t.Errorf("missing branch")
	}
}

func TestPipeViewSelectedNilWhenEmpty(t *testing.T) {
	v := newTestPipeView()
	v.SetSectionData(0, []azdo.Build{})

	if v.SelectedBuild() != nil {
		t.Error("should be nil")
	}
}

func TestPipeViewNoSections(t *testing.T) {
	v := NewPipelineView([]config.PipelineSection{}, newTestTheme())
	v.SetSize(120, 40)

	output := v.View()
	if !strings.Contains(output, "No pipeline sections") {
		t.Errorf("expected no sections: %q", output)
	}
}

func TestPipeViewPreviewContent(t *testing.T) {
	v := newTestPipeView()
	v.SetSize(120, 40)
	v.SetSectionData(0, sampleBuilds())

	output := v.View()
	// Preview should show pipeline name and branch
	if !strings.Contains(output, "CI") {
		t.Error("preview missing pipeline name")
	}
	if !strings.Contains(output, "Alice") {
		t.Error("preview missing triggered by")
	}
}
