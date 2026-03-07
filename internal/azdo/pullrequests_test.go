package azdo

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListPullRequestsCrossRepo(t *testing.T) {
	var gotPath string
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(ListResponse[PullRequest]{
			Count: 1,
			Value: []PullRequest{
				{PullRequestID: 42, Title: "Test PR", Status: "active"},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(server)
	prs, err := c.ListPullRequests("", "", PRSearchCriteria{
		Status:    "active",
		CreatorID: "@me",
	}, 10)

	if err != nil {
		t.Fatal(err)
	}
	if len(prs) != 1 {
		t.Fatalf("got %d PRs, want 1", len(prs))
	}
	if prs[0].PullRequestID != 42 {
		t.Errorf("PR ID = %d, want 42", prs[0].PullRequestID)
	}

	// Verify cross-repo URL (no repository in path)
	if !strings.Contains(gotPath, "/git/pullrequests") {
		t.Errorf("path = %q, want cross-repo endpoint", gotPath)
	}
	if strings.Contains(gotPath, "repositories") {
		t.Errorf("path = %q, should NOT contain 'repositories' for cross-repo", gotPath)
	}

	// Verify query params
	if !strings.Contains(gotQuery, "searchCriteria.status=active") {
		t.Errorf("query missing status filter: %q", gotQuery)
	}
	if !strings.Contains(gotQuery, "searchCriteria.creatorId=user-guid-123") {
		t.Errorf("query should resolve @me to user GUID: %q", gotQuery)
	}
	if !strings.Contains(gotQuery, "%24top=10") {
		t.Errorf("query missing $top: %q", gotQuery)
	}
}

func TestListPullRequestsSingleRepo(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(ListResponse[PullRequest]{Value: []PullRequest{}})
	}))
	defer server.Close()

	c := newTestClient(server)
	c.ListPullRequests("", "", PRSearchCriteria{
		Repository: "my-repo",
	}, 0)

	if !strings.Contains(gotPath, "/git/repositories/my-repo/pullrequests") {
		t.Errorf("path = %q, want single-repo endpoint", gotPath)
	}
}

func TestListPullRequestsOrgOverride(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(ListResponse[PullRequest]{Value: []PullRequest{}})
	}))
	defer server.Close()

	c := newTestClient(server)
	c.ListPullRequests("other-org", "other-project", PRSearchCriteria{}, 0)

	if !strings.Contains(gotPath, "/other-org/other-project/") {
		t.Errorf("path = %q, should use overridden org/project", gotPath)
	}
}

func TestListPullRequestsAllFilters(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(ListResponse[PullRequest]{Value: []PullRequest{}})
	}))
	defer server.Close()

	c := newTestClient(server)
	c.ListPullRequests("", "", PRSearchCriteria{
		Status:     "completed",
		CreatorID:  "specific-user",
		ReviewerID: "@me",
		TargetRef:  "refs/heads/main",
		SourceRef:  "refs/heads/feature",
	}, 5)

	checks := []string{
		"searchCriteria.status=completed",
		"searchCriteria.creatorId=specific-user",
		"searchCriteria.reviewerId=user-guid-123",
		"searchCriteria.targetRefName=refs%2Fheads%2Fmain",
		"searchCriteria.sourceRefName=refs%2Fheads%2Ffeature",
		"%24top=5",
	}
	for _, check := range checks {
		if !strings.Contains(gotQuery, check) {
			t.Errorf("query missing %q in: %q", check, gotQuery)
		}
	}
}

func TestListPullRequestsStatusAll(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(ListResponse[PullRequest]{Value: []PullRequest{}})
	}))
	defer server.Close()

	c := newTestClient(server)
	c.ListPullRequests("", "", PRSearchCriteria{Status: "all"}, 0)

	if strings.Contains(gotQuery, "searchCriteria.status") {
		t.Errorf("status=all should not send filter param: %q", gotQuery)
	}
}

func TestVotePullRequest(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	err := c.VotePullRequest("", "", "repo-id", 42, "reviewer-id", 10)
	if err != nil {
		t.Fatal(err)
	}

	if gotMethod != "PUT" {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if !strings.Contains(gotPath, "/repositories/repo-id/pullrequests/42/reviewers/reviewer-id") {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["vote"] != float64(10) {
		t.Errorf("vote = %v, want 10", gotBody["vote"])
	}
}

func TestUpdatePullRequestStatus(t *testing.T) {
	var gotMethod string
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	err := c.UpdatePullRequestStatus("", "", "repo-id", 42, "completed")
	if err != nil {
		t.Fatal(err)
	}

	if gotMethod != "PATCH" {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotBody["status"] != "completed" {
		t.Errorf("status = %v", gotBody["status"])
	}
}

func TestGetPullRequestIterations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ListResponse[PRIteration]{
			Value: []PRIteration{
				{ID: 1, Description: "first push"},
				{ID: 2, Description: "second push"},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(server)
	iters, err := c.GetPullRequestIterations("", "", "repo-id", 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(iters) != 2 {
		t.Fatalf("got %d iterations, want 2", len(iters))
	}
	if iters[1].ID != 2 {
		t.Errorf("iter[1].ID = %d, want 2", iters[1].ID)
	}
}

func TestGetPullRequestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"message":"Forbidden"}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	_, err := c.ListPullRequests("", "", PRSearchCriteria{}, 0)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
}

func TestGetPullRequestDiff(t *testing.T) {
	var gotPaths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)

		switch {
		case strings.Contains(r.URL.Path, "/iterations") && !strings.Contains(r.URL.Path, "/changes"):
			json.NewEncoder(w).Encode(ListResponse[PRIteration]{
				Value: []PRIteration{{ID: 2}},
			})
		case strings.Contains(r.URL.Path, "/changes"):
			json.NewEncoder(w).Encode(PRChanges{
				ChangeEntries: []PRChangeEntry{{
					ChangeType: "edit",
					Item: PRChangeItem{
						Path:             "/README.md",
						ObjectID:         "new-sha",
						OriginalObjectID: "old-sha",
					},
				}},
			})
		case strings.Contains(r.URL.Path, "/blobs/"):
			if strings.HasSuffix(r.URL.Path, "/old-sha") {
				w.Write([]byte("old\nline\n"))
				return
			}
			if strings.HasSuffix(r.URL.Path, "/new-sha") {
				w.Write([]byte("new\nline\n"))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c := newTestClient(server)
	diff, err := c.GetPullRequestDiff("", "", "repo-id", 42)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(diff, "--- a/README.md") {
		t.Fatalf("diff missing old path header: %q", diff)
	}
	if !strings.Contains(diff, "+++ b/README.md") {
		t.Fatalf("diff missing new path header: %q", diff)
	}
	if !strings.Contains(diff, "-old") {
		t.Fatalf("diff missing removed content: %q", diff)
	}
	if !strings.Contains(diff, "+new") {
		t.Fatalf("diff missing added content: %q", diff)
	}
	if len(gotPaths) != 4 {
		t.Fatalf("got %d requests, want 4", len(gotPaths))
	}
}
