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

func TestLoadMissingProject(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("organization: my-org\nprSections:\n  - title: x\n    filters:\n      status: active\n"), 0644)

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

func TestLoadInvalidPRStatus(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("organization: o\nproject: p\nprSections:\n  - title: x\n    filters:\n      status: bogus\n"), 0644)

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid PR status")
	}
}

func TestLoadMissingSectionTitle(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("organization: o\nproject: p\nprSections:\n  - filters:\n      status: active\n"), 0644)

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for missing section title")
	}
}

func TestLoadWIQLRequired(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("organization: o\nproject: p\nworkItemSections:\n  - title: x\n"), 0644)

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for missing WIQL")
	}
}

func TestLoadCustomBaseURL(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("organization: o\nproject: p\nbaseUrl: https://myserver/collection\nprSections:\n  - title: x\n    filters:\n      status: active\n"), 0644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://myserver/collection" {
		t.Errorf("BaseURL = %q", cfg.BaseURL)
	}
}

func TestLoadCustomLimits(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("organization: o\nproject: p\nprSections:\n  - title: x\n    limit: 50\n    filters:\n      status: active\n"), 0644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PRSections[0].Limit != 50 {
		t.Errorf("Limit = %d, want 50", cfg.PRSections[0].Limit)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("{{invalid yaml"), 0644)

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadPerSectionOrgOverride(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	content := `
organization: default-org
project: default-proj
prSections:
  - title: Other
    organization: other-org
    project: other-proj
    filters:
      status: active
`
	os.WriteFile(cfgPath, []byte(content), 0644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PRSections[0].Organization != "other-org" {
		t.Errorf("section org = %q", cfg.PRSections[0].Organization)
	}
	if cfg.PRSections[0].Project != "other-proj" {
		t.Errorf("section project = %q", cfg.PRSections[0].Project)
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
