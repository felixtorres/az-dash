package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var defaultConfig = Config{
	BaseURL: "https://dev.azure.com",
	Auth: AuthConfig{
		Method: "az-cli",
	},
	PRSections: []PRSection{
		{Title: "Mine", Filters: PRFilters{CreatorID: "@me", Status: "active"}},
		{Title: "Reviewing", Filters: PRFilters{ReviewerID: "@me", Status: "active"}},
		{Title: "All Active", Filters: PRFilters{Status: "active"}},
	},
	WorkItemSections: []WorkItemSection{
		{
			Title: "My Tasks",
			WIQL:  "SELECT [System.Id] FROM WorkItems WHERE [System.AssignedTo] = @me AND [System.State] <> 'Closed' ORDER BY [System.ChangedDate] DESC",
		},
		{
			Title: "Bugs",
			WIQL:  "SELECT [System.Id] FROM WorkItems WHERE [System.WorkItemType] = 'Bug' AND [System.State] <> 'Closed' ORDER BY [Microsoft.VSTS.Common.Priority]",
		},
	},
	PipelineSections: []PipelineSection{
		{Title: "My Runs", Filters: PipelineFilters{RequestedFor: "@me"}},
		{Title: "Failed", Filters: PipelineFilters{ResultFilter: "failed"}},
		{Title: "All Recent"},
	},
	Defaults: Defaults{
		View:                   "prs",
		RefetchIntervalMinutes: 5,
		Preview:                PreviewConfig{Open: true, Width: 80},
		PRsLimit:               20,
		WorkItemsLimit:         20,
		PipelinesLimit:         20,
	},
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return generateDefault(path)
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := defaultConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	applyDefaults(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func generateDefault(path string) (*Config, error) {
	cfg := defaultConfig

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling default config: %w", err)
	}

	header := "# az-dash configuration\n# See: https://github.com/felixtorres/az-dash\n\n"
	if err := os.WriteFile(path, []byte(header+string(data)), 0644); err != nil {
		return nil, fmt.Errorf("writing default config to %s: %w", path, err)
	}

	fmt.Printf("Created default config at %s\n", path)
	fmt.Println("Edit it to set your organization and project, then re-run az-dash.")
	os.Exit(0)

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://dev.azure.com"
	}
	if cfg.Auth.Method == "" {
		cfg.Auth.Method = "az-cli"
	}
	if cfg.Defaults.View == "" {
		cfg.Defaults.View = "prs"
	}
	if cfg.Defaults.RefetchIntervalMinutes == 0 {
		cfg.Defaults.RefetchIntervalMinutes = 5
	}
	if cfg.Defaults.Preview.Width == 0 {
		cfg.Defaults.Preview.Width = 80
	}
	if cfg.Defaults.PRsLimit == 0 {
		cfg.Defaults.PRsLimit = 20
	}
	if cfg.Defaults.WorkItemsLimit == 0 {
		cfg.Defaults.WorkItemsLimit = 20
	}
	if cfg.Defaults.PipelinesLimit == 0 {
		cfg.Defaults.PipelinesLimit = 20
	}

	for i := range cfg.PRSections {
		if cfg.PRSections[i].Limit == 0 {
			cfg.PRSections[i].Limit = cfg.Defaults.PRsLimit
		}
	}
	for i := range cfg.WorkItemSections {
		if cfg.WorkItemSections[i].Limit == 0 {
			cfg.WorkItemSections[i].Limit = cfg.Defaults.WorkItemsLimit
		}
	}
	for i := range cfg.PipelineSections {
		if cfg.PipelineSections[i].Limit == 0 {
			cfg.PipelineSections[i].Limit = cfg.Defaults.PipelinesLimit
		}
	}
}
