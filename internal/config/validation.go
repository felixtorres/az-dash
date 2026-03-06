package config

import (
	"fmt"
	"strings"
)

func validate(cfg *Config) error {
	if cfg.Organization == "" {
		return fmt.Errorf("'organization' is required in config")
	}
	if cfg.Project == "" {
		return fmt.Errorf("'project' is required in config")
	}

	if cfg.Auth.Method != "az-cli" && cfg.Auth.Method != "pat" {
		return fmt.Errorf("auth.method must be 'az-cli' or 'pat', got %q", cfg.Auth.Method)
	}
	if cfg.Auth.Method == "pat" && cfg.Auth.PAT == "" {
		return fmt.Errorf("auth.pat is required when auth.method is 'pat'")
	}

	validViews := map[string]bool{"prs": true, "workitems": true, "pipelines": true}
	if !validViews[cfg.Defaults.View] {
		return fmt.Errorf("defaults.view must be 'prs', 'workitems', or 'pipelines', got %q", cfg.Defaults.View)
	}

	for i, s := range cfg.PRSections {
		if s.Title == "" {
			return fmt.Errorf("prSections[%d]: title is required", i)
		}
		if s.Filters.Status != "" {
			valid := map[string]bool{"active": true, "completed": true, "abandoned": true, "all": true}
			if !valid[strings.ToLower(s.Filters.Status)] {
				return fmt.Errorf("prSections[%d]: invalid status %q", i, s.Filters.Status)
			}
		}
	}

	for i, s := range cfg.WorkItemSections {
		if s.Title == "" {
			return fmt.Errorf("workItemSections[%d]: title is required", i)
		}
		if s.WIQL == "" {
			return fmt.Errorf("workItemSections[%d]: wiql is required", i)
		}
	}

	for i, s := range cfg.PipelineSections {
		if s.Title == "" {
			return fmt.Errorf("pipelineSections[%d]: title is required", i)
		}
	}

	return nil
}
