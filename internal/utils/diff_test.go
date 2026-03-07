package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/felixtorres/az-dash/internal/azdo"
)

func TestDiffCommandUsesDeltaWithStdin(t *testing.T) {
	if _, err := exec.LookPath("delta"); err != nil {
		t.Skip("delta not installed")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/iterations") && !strings.Contains(r.URL.Path, "/changes"):
			json.NewEncoder(w).Encode(azdo.ListResponse[azdo.PRIteration]{
				Value: []azdo.PRIteration{{ID: 2}},
			})
		case strings.Contains(r.URL.Path, "/changes"):
			json.NewEncoder(w).Encode(azdo.PRChanges{
				ChangeEntries: []azdo.PRChangeEntry{{
					ChangeType: "edit",
					Item: azdo.PRChangeItem{
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

	cmd := DiffCommand(testClient(server.URL), "repo-id", 42)
	if filepath.Base(cmd.Path) != "sh" {
		t.Fatalf("expected shell wrapper command, got %q", cmd.Path)
	}
	if len(cmd.Args) < 3 || cmd.Args[1] != "-c" {
		t.Fatalf("expected shell command args, got %v", cmd.Args)
	}
	script := cmd.Args[2]
	if !strings.Contains(script, "delta") {
		t.Fatalf("expected script to invoke delta, got %q", script)
	}
	if !strings.Contains(script, "less -R") {
		t.Fatalf("expected script to fall back to less, got %q", script)
	}

	path := extractQuotedPath(script)
	if path == "" {
		t.Fatalf("expected temp diff file path in script, got %q", script)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "--- a/README.md") {
		t.Fatalf("diff file missing expected header: %q", string(content))
	}
	if !strings.Contains(string(content), "+++ b/README.md") {
		t.Fatalf("diff file missing expected header: %q", string(content))
	}
	if !strings.Contains(string(content), "-old") {
		t.Fatalf("diff file missing expected removed line: %q", string(content))
	}
	if !strings.Contains(string(content), "+new") {
		t.Fatalf("diff file missing expected added line: %q", string(content))
	}
	if cmd.Stdout == nil || cmd.Stderr == nil {
		t.Fatal("expected stdout and stderr to be configured")
	}
	if cmd.Stdin == nil {
		t.Fatal("expected stdin to be configured")
	}
}

func testClient(baseURL string) *azdo.Client {
	auth := &testAuth{token: "test-token"}
	c := azdo.NewClient(baseURL, "test-org", "test-project", auth)
	return c
}

type testAuth struct {
	token string
}

func (m *testAuth) GetToken() (string, error) {
	return m.token, nil
}

func (m *testAuth) AuthHeader() (string, string, error) {
	return "Authorization", "Bearer " + m.token, nil
}

func extractQuotedPath(script string) string {
	parts := strings.Split(script, "\"")
	for _, part := range parts {
		if strings.Contains(part, "az-dash-diff-") {
			return part
		}
	}
	return ""
}
