package views

import (
	"fmt"
	"strings"
	"testing"

	"github.com/felixtorres/az-dash/internal/azdo"
	"github.com/felixtorres/az-dash/internal/config"
	"github.com/felixtorres/az-dash/internal/tui/theme"
)

func newTestTheme() *theme.Theme {
	return theme.New(config.Theme{})
}

func newTestPRView() *PRView {
	return NewPRView([]config.PRSection{
		{Title: "Mine"},
		{Title: "Reviewing"},
		{Title: "All"},
	}, newTestTheme())
}

func samplePRs() []azdo.PullRequest {
	return []azdo.PullRequest{
		{PullRequestID: 1, Title: "Fix auth bug", Status: "active",
			CreatedBy:     azdo.IdentityRef{DisplayName: "Alice"},
			SourceRefName: "refs/heads/fix-auth", TargetRefName: "refs/heads/main",
			Reviewers: []azdo.Reviewer{
				{DisplayName: "Bob", Vote: 10},
				{DisplayName: "Carol", Vote: 0},
			}},
		{PullRequestID: 2, Title: "Add feature", Status: "active", IsDraft: true,
			CreatedBy: azdo.IdentityRef{DisplayName: "Bob"}},
		{PullRequestID: 3, Title: "Refactor", Status: "completed",
			CreatedBy: azdo.IdentityRef{DisplayName: "Carol"}},
	}
}

func TestPRViewSectionNavigation(t *testing.T) {
	v := newTestPRView()

	if v.ActiveSectionIndex() != 0 {
		t.Errorf("initial section = %d, want 0", v.ActiveSectionIndex())
	}

	v.NextSection()
	if v.ActiveSectionIndex() != 1 {
		t.Errorf("after NextSection = %d, want 1", v.ActiveSectionIndex())
	}

	v.NextSection()
	if v.ActiveSectionIndex() != 2 {
		t.Errorf("after 2x NextSection = %d, want 2", v.ActiveSectionIndex())
	}

	v.NextSection() // wraps
	if v.ActiveSectionIndex() != 0 {
		t.Errorf("after wrap = %d, want 0", v.ActiveSectionIndex())
	}

	v.PrevSection() // wraps backwards
	if v.ActiveSectionIndex() != 2 {
		t.Errorf("after PrevSection wrap = %d, want 2", v.ActiveSectionIndex())
	}
}

func TestPRViewCursorNavigation(t *testing.T) {
	v := newTestPRView()
	v.SetSize(120, 40)
	v.SetSectionData(0, samplePRs())

	// Initial cursor at 0
	if pr := v.SelectedPR(); pr.PullRequestID != 1 {
		t.Errorf("initial selected PR = %d, want 1", pr.PullRequestID)
	}

	v.CursorDown()
	if pr := v.SelectedPR(); pr.PullRequestID != 2 {
		t.Errorf("after CursorDown = %d, want 2", pr.PullRequestID)
	}

	v.CursorDown()
	if pr := v.SelectedPR(); pr.PullRequestID != 3 {
		t.Errorf("after 2x CursorDown = %d, want 3", pr.PullRequestID)
	}

	v.CursorDown() // at end, should not move
	if pr := v.SelectedPR(); pr.PullRequestID != 3 {
		t.Errorf("CursorDown at end = %d, want 3", pr.PullRequestID)
	}

	v.CursorUp()
	if pr := v.SelectedPR(); pr.PullRequestID != 2 {
		t.Errorf("after CursorUp = %d, want 2", pr.PullRequestID)
	}

	v.CursorFirst()
	if pr := v.SelectedPR(); pr.PullRequestID != 1 {
		t.Errorf("after CursorFirst = %d, want 1", pr.PullRequestID)
	}

	v.CursorLast()
	if pr := v.SelectedPR(); pr.PullRequestID != 3 {
		t.Errorf("after CursorLast = %d, want 3", pr.PullRequestID)
	}
}

func TestPRViewCursorUpAtZero(t *testing.T) {
	v := newTestPRView()
	v.SetSectionData(0, samplePRs())

	v.CursorUp() // already at 0
	if pr := v.SelectedPR(); pr.PullRequestID != 1 {
		t.Errorf("CursorUp at 0 = %d, want 1", pr.PullRequestID)
	}
}

func TestPRViewSectionChangeCursorReset(t *testing.T) {
	v := newTestPRView()
	v.SetSectionData(0, samplePRs())
	v.SetSectionData(1, []azdo.PullRequest{
		{PullRequestID: 10, Title: "Review PR"},
	})

	v.CursorDown()
	v.CursorDown()
	// Cursor at 2 in section 0
	v.NextSection()
	// Cursor should reset to 0 in section 1
	if pr := v.SelectedPR(); pr.PullRequestID != 10 {
		t.Errorf("after section change, PR = %d, want 10", pr.PullRequestID)
	}
}

func TestPRViewTogglePreview(t *testing.T) {
	v := newTestPRView()
	v.SetSize(120, 40)
	v.SetSectionData(0, samplePRs())

	// Preview on by default
	output := v.View()
	if !strings.Contains(output, "│") {
		t.Error("preview pane separator missing with preview on")
	}

	v.TogglePreview()
	output = v.View()
	if strings.Contains(output, "│") {
		t.Error("preview pane separator present with preview off")
	}
}

func TestPRViewLoadingState(t *testing.T) {
	v := newTestPRView()
	v.SetSize(120, 40)
	// Section 0 is still in loading state (no data set)

	output := v.View()
	if !strings.Contains(output, "Loading PRs") {
		t.Errorf("expected loading message, got: %q", output)
	}
}

func TestPRViewErrorState(t *testing.T) {
	v := newTestPRView()
	v.SetSize(120, 40)
	v.SetSectionError(0, fmt.Errorf("auth failed"))

	output := v.View()
	if !strings.Contains(output, "auth failed") {
		t.Errorf("expected error message, got: %q", output)
	}
}

func TestPRViewEmptyState(t *testing.T) {
	v := newTestPRView()
	v.SetSize(120, 40)
	v.SetSectionData(0, []azdo.PullRequest{})

	output := v.View()
	if !strings.Contains(output, "No pull requests") {
		t.Errorf("expected empty message, got: %q", output)
	}
}

func TestPRViewNoSections(t *testing.T) {
	v := NewPRView([]config.PRSection{}, newTestTheme())
	v.SetSize(120, 40)

	output := v.View()
	if !strings.Contains(output, "No PR sections") {
		t.Errorf("expected no sections message, got: %q", output)
	}
}

func TestPRViewSelectedPRNilWhenEmpty(t *testing.T) {
	v := newTestPRView()
	v.SetSectionData(0, []azdo.PullRequest{})

	if v.SelectedPR() != nil {
		t.Error("SelectedPR should be nil when empty")
	}
}

func TestPRViewSetSectionDataOutOfBounds(t *testing.T) {
	v := newTestPRView()
	// Should not panic
	v.SetSectionData(99, samplePRs())
	v.SetSectionError(99, fmt.Errorf("test"))
}

func TestPRViewSectionTabCounts(t *testing.T) {
	v := newTestPRView()
	v.SetSize(120, 40)
	v.SetSectionData(0, samplePRs())
	v.SetSectionData(1, []azdo.PullRequest{})

	output := v.View()
	if !strings.Contains(output, "Mine (3)") {
		t.Errorf("expected 'Mine (3)' in output")
	}
	if !strings.Contains(output, "Reviewing (0)") {
		t.Errorf("expected 'Reviewing (0)' in output")
	}
}

func TestVoteString(t *testing.T) {
	tests := []struct {
		vote int
		want string
	}{
		{10, "✓"},
		{5, "~"},
		{-5, "…"},
		{-10, "✗"},
		{0, "○"},
		{99, "○"},
	}
	for _, tt := range tests {
		if got := voteString(tt.vote); got != tt.want {
			t.Errorf("voteString(%d) = %q, want %q", tt.vote, got, tt.want)
		}
	}
}

func TestPRPreviewMergeStatusDisplay(t *testing.T) {
	v := newTestPRView()
	v.SetSize(120, 40)
	v.SetSectionData(0, []azdo.PullRequest{{
		PullRequestID: 1,
		Title:         "Test PR",
		Status:        "active",
		MergeStatus:   "succeeded",
		CreatedBy:     azdo.IdentityRef{DisplayName: "Alice"},
		SourceRefName: "refs/heads/feature",
		TargetRefName: "refs/heads/main",
	}})

	output := v.View()
	if !strings.Contains(output, "Mergeability") {
		t.Fatalf("expected mergeability label, got: %q", output)
	}
	if !strings.Contains(output, "mergeable") {
		t.Fatalf("expected friendly merge status, got: %q", output)
	}
	if strings.Contains(output, "Merge Status") {
		t.Fatalf("expected old merge status label to be absent, got: %q", output)
	}
}

func TestMergeStatusDisplay(t *testing.T) {
	tests := []struct {
		status    string
		wantLabel string
		wantIcon  string
	}{
		{"succeeded", "mergeable", "✓"},
		{"conflicts", "conflicts", "✗"},
		{"queued", "checking", "◌"},
		{"rejectedByPolicy", "blocked by policy", "!"},
		{"failure", "merge failed", "✗"},
		{"notSet", "unknown", "⊘"},
	}

	for _, tt := range tests {
		label, icon := mergeStatusDisplay(tt.status)
		if label != tt.wantLabel || icon != tt.wantIcon {
			t.Fatalf("mergeStatusDisplay(%q) = (%q, %q), want (%q, %q)", tt.status, label, icon, tt.wantLabel, tt.wantIcon)
		}
	}
}

func TestTrimRef(t *testing.T) {
	if got := trimRef("refs/heads/main"); got != "main" {
		t.Errorf("trimRef = %q", got)
	}
	if got := trimRef("main"); got != "main" {
		t.Errorf("trimRef = %q", got)
	}
}
