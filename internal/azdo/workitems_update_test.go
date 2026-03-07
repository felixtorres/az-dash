package azdo

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateWorkItemField(t *testing.T) {
	var gotMethod, gotContentType string
	var gotBody []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	err := c.UpdateWorkItemField("", "", 42, "System.State", "Active")
	if err != nil {
		t.Fatal(err)
	}

	if gotMethod != "PATCH" {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotContentType != "application/json-patch+json" {
		t.Errorf("Content-Type = %q, want application/json-patch+json", gotContentType)
	}
	if len(gotBody) != 1 {
		t.Fatalf("patch ops = %d, want 1", len(gotBody))
	}
	if gotBody[0]["op"] != "replace" {
		t.Errorf("op = %v", gotBody[0]["op"])
	}
	if gotBody[0]["path"] != "/fields/System.State" {
		t.Errorf("path = %v", gotBody[0]["path"])
	}
	if gotBody[0]["value"] != "Active" {
		t.Errorf("value = %v", gotBody[0]["value"])
	}
}

func TestUpdateWorkItemFieldAssign(t *testing.T) {
	var gotBody []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	c.UpdateWorkItemField("", "", 100, "System.AssignedTo", "John Doe")

	if gotBody[0]["path"] != "/fields/System.AssignedTo" {
		t.Errorf("path = %v", gotBody[0]["path"])
	}
	if gotBody[0]["value"] != "John Doe" {
		t.Errorf("value = %v", gotBody[0]["value"])
	}
}

func TestUpdateWorkItemFieldUnassign(t *testing.T) {
	var gotBody []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	c.UpdateWorkItemField("", "", 100, "System.AssignedTo", "")

	if gotBody[0]["value"] != "" {
		t.Errorf("value = %v, want empty string for unassign", gotBody[0]["value"])
	}
}
