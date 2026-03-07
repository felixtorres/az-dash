package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")

	content := `
organization: my-org
project: my-project
auth:
  method: pat
  pat: test-token
prSections:
  - title: Mine
    filters:
      creatorId: "@me"
      status: active
workItemSections:
  - title: Tasks
    wiql: "SELECT [System.Id] FROM WorkItems WHERE [System.AssignedTo] = @me"
pipelineSections:
  - title: All
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Organization != "my-org" {
		t.Errorf("Organization = %q, want %q", cfg.Organization, "my-org")
	}
	if cfg.Project != "my-project" {
		t.Errorf("Project = %q, want %q", cfg.Project, "my-project")
	}
	if cfg.Auth.Method != "pat" {
		t.Errorf("Auth.Method = %q, want %q", cfg.Auth.Method, "pat")
	}
	if cfg.BaseURL != "https://dev.azure.com" {
		t.Errorf("BaseURL = %q, want default", cfg.BaseURL)
	}
	if len(cfg.PRSections) != 1 {
		t.Errorf("PRSections count = %d, want 1", len(cfg.PRSections))
	}
	if cfg.PRSections[0].Limit != 20 {
		t.Errorf("PRSections[0].Limit = %d, want 20 (default)", cfg.PRSections[0].Limit)
	}
}

func TestLoadMissingOrg(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")

	content := `
project: my-project
prSections:
  - title: Mine
    filters:
      status: active
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for missing organization")
	}
}

func TestLoadInvalidAuthMethod(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")

	content := `
organization: my-org
project: my-project
auth:
  method: oauth
prSections:
  - title: Mine
    filters:
      status: active
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid auth method")
	}
}

func TestLoadPATRequiredWhenMethodIsPAT(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")

	content := `
organization: my-org
project: my-project
auth:
  method: pat
prSections:
  - title: Mine
    filters:
      status: active
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for missing PAT")
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := Config{
		Organization: "org",
		Project:      "proj",
		PRSections: []PRSection{
			{Title: "Test"},
		},
		WorkItemSections: []WorkItemSection{
			{Title: "Test", WIQL: "SELECT 1"},
		},
	}

	applyDefaults(&cfg)

	if cfg.BaseURL != "https://dev.azure.com" {
		t.Errorf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.Defaults.RefetchIntervalMinutes != 5 {
		t.Errorf("RefetchIntervalMinutes = %d", cfg.Defaults.RefetchIntervalMinutes)
	}
	if cfg.PRSections[0].Limit != 20 {
		t.Errorf("PRSection limit = %d", cfg.PRSections[0].Limit)
	}
}
