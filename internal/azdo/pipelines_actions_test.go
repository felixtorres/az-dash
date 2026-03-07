package azdo

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCancelBuild(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	err := c.CancelBuild("", "", 42)
	if err != nil {
		t.Fatal(err)
	}

	if gotMethod != "PATCH" {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if !strings.Contains(gotPath, "/build/builds/42") {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["status"] != "cancelling" {
		t.Errorf("status = %v", gotBody["status"])
	}
}

func TestGetBuildLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/logs") {
			t.Errorf("path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ListResponse[BuildLog]{
			Value: []BuildLog{
				{ID: 1, LineCount: 50},
				{ID: 2, LineCount: 100},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(server)
	logs, err := c.GetBuildLogs("", "", 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("got %d logs, want 2", len(logs))
	}
	if logs[1].LineCount != 100 {
		t.Errorf("log[1].LineCount = %d", logs[1].LineCount)
	}
}

func TestGetBuildLogContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/logs/1") {
			t.Errorf("path = %q", r.URL.Path)
		}
		accept := r.Header.Get("Accept")
		if accept != "text/plain" {
			t.Errorf("Accept = %q, want text/plain", accept)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Starting build...\nBuild succeeded."))
	}))
	defer server.Close()

	c := newTestClient(server)
	content, err := c.GetBuildLogContent("", "", 42, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, "Build succeeded") {
		t.Errorf("content = %q", content)
	}
}

func TestGetBuildLogContentError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetBuildLogContent("", "", 42, 999)
	if err == nil {
		t.Fatal("expected error")
	}
}
