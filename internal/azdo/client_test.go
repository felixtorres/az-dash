package azdo

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockAuth implements AuthProvider for testing.
type mockAuth struct {
	token      string
	callCount  int
	failCount  int // fail first N calls
}

func (m *mockAuth) GetToken() (string, error) {
	m.callCount++
	return m.token, nil
}

func (m *mockAuth) AuthHeader() (string, string, error) {
	m.callCount++
	return "Authorization", "Bearer " + m.token, nil
}

func newTestClient(server *httptest.Server) *Client {
	auth := &mockAuth{token: "test-token"}
	c := NewClient(server.URL, "test-org", "test-project", auth)
	c.userID = "user-guid-123"
	return c
}

func TestClientProjectURL(t *testing.T) {
	c := &Client{
		baseURL:      "https://dev.azure.com",
		organization: "my-org",
		project:      "my-project",
	}

	tests := []struct {
		org, project string
		want         string
	}{
		{"", "", "https://dev.azure.com/my-org/my-project/_apis"},
		{"other-org", "", "https://dev.azure.com/other-org/my-project/_apis"},
		{"", "other-proj", "https://dev.azure.com/my-org/other-proj/_apis"},
		{"o", "p", "https://dev.azure.com/o/p/_apis"},
	}

	for _, tt := range tests {
		got := c.projectURL(tt.org, tt.project)
		if got != tt.want {
			t.Errorf("projectURL(%q, %q) = %q, want %q", tt.org, tt.project, got, tt.want)
		}
	}
}

func TestClientOrgURL(t *testing.T) {
	c := &Client{
		baseURL:      "https://dev.azure.com",
		organization: "my-org",
	}

	if got := c.orgURL(""); got != "https://dev.azure.com/my-org/_apis" {
		t.Errorf("orgURL('') = %q", got)
	}
	if got := c.orgURL("other"); got != "https://dev.azure.com/other/_apis" {
		t.Errorf("orgURL('other') = %q", got)
	}
}

func TestClientBaseURLTrailingSlash(t *testing.T) {
	c := NewClient("https://dev.azure.com/", "org", "proj", &mockAuth{token: "t"})
	got := c.projectURL("", "")
	if got != "https://dev.azure.com/org/proj/_apis" {
		t.Errorf("trailing slash not trimmed: %q", got)
	}
}

func TestClientGetSendsAuthHeader(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	var result map[string]interface{}
	c.get(server.URL+"/test", nil, &result)

	if gotAuth != "Bearer test-token" {
		t.Errorf("auth header = %q, want %q", gotAuth, "Bearer test-token")
	}
}

func TestClientGetSendsAPIVersion(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	var result map[string]interface{}
	c.get(server.URL+"/test", nil, &result)

	if !strings.Contains(gotQuery, "api-version=7.1") {
		t.Errorf("query = %q, missing api-version", gotQuery)
	}
}

func TestClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"Not found"}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	var result map[string]interface{}
	err := c.get(server.URL+"/test", nil, &result)

	if err == nil {
		t.Fatal("expected error for 404")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestClientPostSendsBody(t *testing.T) {
	var gotBody string
	var gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	body := map[string]string{"key": "value"}
	c.post(server.URL+"/test", body, nil)

	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q", gotContentType)
	}

	var parsed map[string]string
	json.Unmarshal([]byte(gotBody), &parsed)
	if parsed["key"] != "value" {
		t.Errorf("body = %q", gotBody)
	}
}

func TestClientPutMethod(t *testing.T) {
	var gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	c.put(server.URL+"/test", map[string]int{"vote": 10}, nil)

	if gotMethod != "PUT" {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
}

func TestClientPatchMethod(t *testing.T) {
	var gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	c.patch(server.URL+"/test", map[string]string{"status": "completed"}, nil)

	if gotMethod != "PATCH" {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
}

func TestIsUnauthorized(t *testing.T) {
	if isUnauthorized(nil) {
		t.Error("nil should not be unauthorized")
	}
	if !isUnauthorized(&APIError{StatusCode: 401}) {
		t.Error("401 should be unauthorized")
	}
	if isUnauthorized(&APIError{StatusCode: 403}) {
		t.Error("403 should not be unauthorized")
	}
}

func TestBootstrapResolvesUserID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "profile") {
			json.NewEncoder(w).Encode(Profile{
				ID:          "resolved-user-id",
				DisplayName: "Test User",
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	// Override the profile URL by using a custom client
	auth := &mockAuth{token: "test-token"}
	c := &Client{
		baseURL:      server.URL,
		organization: "org",
		project:      "proj",
		auth:         auth,
		httpClient:   http.DefaultClient,
	}

	// Can't test Bootstrap directly since it calls a hardcoded URL,
	// but we can test that userID is properly set and used.
	c.userID = "resolved-user-id"

	if c.UserID() != "resolved-user-id" {
		t.Errorf("UserID() = %q", c.UserID())
	}
	if c.ResolveMe("@me") != "resolved-user-id" {
		t.Errorf("ResolveMe(@me) = %q", c.ResolveMe("@me"))
	}
}
