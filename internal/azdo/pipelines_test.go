package azdo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestListBuilds(t *testing.T) {
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		now := time.Now()
		json.NewEncoder(w).Encode(ListResponse[Build]{
			Value: []Build{
				{
					ID:           1,
					BuildNumber:  "20260306.1",
					Status:       "completed",
					Result:       "succeeded",
					SourceBranch: "refs/heads/main",
					FinishTime:   &now,
					Definition:   BuildDefinition{ID: 10, Name: "CI Pipeline"},
					RequestedFor: IdentityRef{DisplayName: "Test User"},
				},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(server)
	builds, err := c.ListBuilds("", "", BuildSearchCriteria{
		DefinitionID: 10,
		RequestedFor: "@me",
		ResultFilter: "succeeded",
		BranchName:   "main",
	}, 5)

	if err != nil {
		t.Fatal(err)
	}
	if len(builds) != 1 {
		t.Fatalf("got %d builds, want 1", len(builds))
	}
	if builds[0].BuildNumber != "20260306.1" {
		t.Errorf("BuildNumber = %q", builds[0].BuildNumber)
	}

	// Verify the builds API path
	if !strings.Contains(gotPath, "/build/builds") {
		t.Errorf("path = %q, want /build/builds", gotPath)
	}

	// Verify query params (values are URL-encoded)
	checks := map[string]string{
		"definitions":  "10",
		"requestedFor": "user-guid-123",
		"resultFilter": "succeeded",
		"branchName":   "refs%2Fheads%2Fmain",
	}
	for param, expected := range checks {
		if !strings.Contains(gotQuery, param+"="+expected) {
			t.Errorf("query missing %s=%s in: %q", param, expected, gotQuery)
		}
	}
	if !strings.Contains(gotQuery, "%24top=5") {
		t.Errorf("query missing $top: %q", gotQuery)
	}
}

func TestListBuildsBranchPrefixing(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"main", "refs/heads/main"},
		{"refs/heads/main", "refs/heads/main"},
		{"refs/pull/42/merge", "refs/pull/42/merge"},
		{"feature/foo", "refs/heads/feature/foo"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var gotQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotQuery = r.URL.RawQuery
				json.NewEncoder(w).Encode(ListResponse[Build]{Value: []Build{}})
			}))
			defer server.Close()

			c := newTestClient(server)
			c.ListBuilds("", "", BuildSearchCriteria{BranchName: tt.input}, 0)

			if !strings.Contains(gotQuery, "branchName="+strings.ReplaceAll(tt.want, "/", "%2F")) {
				t.Errorf("for input %q, query = %q, want branchName=%s", tt.input, gotQuery, tt.want)
			}
		})
	}
}

func TestListBuildsNoFilters(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(ListResponse[Build]{Value: []Build{}})
	}))
	defer server.Close()

	c := newTestClient(server)
	c.ListBuilds("", "", BuildSearchCriteria{}, 0)

	// Should only have api-version
	params := strings.Split(gotQuery, "&")
	for _, p := range params {
		if !strings.HasPrefix(p, "api-version") {
			t.Errorf("unexpected param with no filters: %q", p)
		}
	}
}

func TestBuildDuration(t *testing.T) {
	start := time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC)
	finish := time.Date(2026, 3, 6, 10, 5, 30, 0, time.UTC)

	b := Build{StartTime: &start, FinishTime: &finish}
	d := b.Duration()
	if d != 5*time.Minute+30*time.Second {
		t.Errorf("Duration() = %v, want 5m30s", d)
	}
}

func TestBuildDurationNilTimes(t *testing.T) {
	b := Build{}
	if b.Duration() != 0 {
		t.Errorf("Duration() with nil times = %v, want 0", b.Duration())
	}
}

func TestGetBuildTimeline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/timeline") {
			t.Errorf("expected timeline path, got %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Timeline{
			Records: []TimelineRecord{
				{ID: "1", Name: "Build", Type: "Stage", State: "completed", Result: "succeeded"},
				{ID: "2", ParentID: "1", Name: "Build Job", Type: "Job", State: "completed"},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(server)
	timeline, err := c.GetBuildTimeline("", "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(timeline.Records) != 2 {
		t.Fatalf("got %d records, want 2", len(timeline.Records))
	}
	if timeline.Records[0].Type != "Stage" {
		t.Errorf("record[0].Type = %q", timeline.Records[0].Type)
	}
}
