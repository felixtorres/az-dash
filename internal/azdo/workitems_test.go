package azdo

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQueryWorkItems(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Method == "POST" && strings.Contains(r.URL.Path, "wiql") {
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			if req["query"] != "SELECT [System.Id] FROM WorkItems WHERE 1=1" {
				t.Errorf("unexpected WIQL: %v", req["query"])
			}

			json.NewEncoder(w).Encode(WIQLResult{
				WorkItems: []WIQLWorkItemRef{
					{ID: 100}, {ID: 200},
				},
			})
			return
		}
		if r.Method == "GET" && strings.Contains(r.URL.Path, "workitems") {
			ids := r.URL.Query().Get("ids")
			if ids != "100,200" {
				t.Errorf("ids param = %q, want '100,200'", ids)
			}
			expand := r.URL.Query().Get("$expand")
			if expand != "all" {
				t.Errorf("$expand = %q, want 'all'", expand)
			}

			json.NewEncoder(w).Encode(ListResponse[WorkItem]{
				Value: []WorkItem{
					{ID: 100, Fields: map[string]interface{}{"System.Title": "Bug fix"}},
					{ID: 200, Fields: map[string]interface{}{"System.Title": "Feature"}},
				},
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(404)
	}))
	defer server.Close()

	c := newTestClient(server)
	items, err := c.QueryWorkItems("", "", "SELECT [System.Id] FROM WorkItems WHERE 1=1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].StringField("System.Title") != "Bug fix" {
		t.Errorf("item[0].Title = %q", items[0].StringField("System.Title"))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (wiql + workitems), got %d", callCount)
	}
}

func TestQueryWorkItemsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(WIQLResult{WorkItems: []WIQLWorkItemRef{}})
	}))
	defer server.Close()

	c := newTestClient(server)
	items, err := c.QueryWorkItems("", "", "SELECT 1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if items != nil {
		t.Errorf("expected nil for empty result, got %v", items)
	}
}

func TestQueryWorkItemsBatching(t *testing.T) {
	batchCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// Return 250 work item IDs
			refs := make([]WIQLWorkItemRef, 250)
			for i := range refs {
				refs[i] = WIQLWorkItemRef{ID: i + 1}
			}
			json.NewEncoder(w).Encode(WIQLResult{WorkItems: refs})
			return
		}
		if r.Method == "GET" {
			batchCalls++
			ids := r.URL.Query().Get("ids")
			idList := strings.Split(ids, ",")

			items := make([]WorkItem, len(idList))
			for i := range items {
				items[i] = WorkItem{ID: i + 1, Fields: map[string]interface{}{}}
			}
			json.NewEncoder(w).Encode(ListResponse[WorkItem]{Value: items})
			return
		}
	}))
	defer server.Close()

	c := newTestClient(server)
	items, err := c.QueryWorkItems("", "", "SELECT 1", 0)
	if err != nil {
		t.Fatal(err)
	}

	// 250 IDs should result in 2 batch calls (200 + 50)
	if batchCalls != 2 {
		t.Errorf("expected 2 batch calls for 250 items, got %d", batchCalls)
	}
	if len(items) != 250 {
		t.Errorf("got %d items, want 250", len(items))
	}
}

func TestWorkItemStringField(t *testing.T) {
	wi := WorkItem{
		ID: 1,
		Fields: map[string]interface{}{
			"System.Title":      "My Title",
			"System.AssignedTo": map[string]interface{}{"displayName": "John Doe", "id": "123"},
			"System.State":      "Active",
		},
	}

	tests := []struct {
		field string
		want  string
	}{
		{"System.Title", "My Title"},
		{"System.AssignedTo", "John Doe"},
		{"System.State", "Active"},
		{"System.Missing", ""},
	}

	for _, tt := range tests {
		got := wi.StringField(tt.field)
		if got != tt.want {
			t.Errorf("StringField(%q) = %q, want %q", tt.field, got, tt.want)
		}
	}
}
