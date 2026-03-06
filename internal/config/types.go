package config

type Config struct {
	Organization string             `yaml:"organization"`
	Project      string             `yaml:"project"`
	BaseURL      string             `yaml:"baseUrl"`
	Auth         AuthConfig         `yaml:"auth"`
	PRSections   []PRSection        `yaml:"prSections"`
	WorkItemSections []WorkItemSection `yaml:"workItemSections"`
	PipelineSections []PipelineSection `yaml:"pipelineSections"`
	Defaults     Defaults           `yaml:"defaults"`
	Theme        Theme              `yaml:"theme"`
	Keybindings  KeybindingsConfig  `yaml:"keybindings"`
}

type AuthConfig struct {
	Method string `yaml:"method"` // "pat" or "az-cli"
	PAT    string `yaml:"pat"`
}

type PRSection struct {
	Title        string          `yaml:"title"`
	Organization string          `yaml:"organization"`
	Project      string          `yaml:"project"`
	Filters      PRFilters       `yaml:"filters"`
	Limit        int             `yaml:"limit"`
	Layout       map[string]ColumnLayout `yaml:"layout"`
}

type PRFilters struct {
	Status        string `yaml:"status"`
	CreatorID     string `yaml:"creatorId"`
	ReviewerID    string `yaml:"reviewerId"`
	Repository    string `yaml:"repository"`
	TargetBranch  string `yaml:"targetBranch"`
	SourceRefName string `yaml:"sourceRefName"`
}

type WorkItemSection struct {
	Title        string          `yaml:"title"`
	Organization string          `yaml:"organization"`
	Project      string          `yaml:"project"`
	WIQL         string          `yaml:"wiql"`
	Limit        int             `yaml:"limit"`
	Layout       map[string]ColumnLayout `yaml:"layout"`
}

type PipelineSection struct {
	Title        string          `yaml:"title"`
	Organization string          `yaml:"organization"`
	Project      string          `yaml:"project"`
	Filters      PipelineFilters `yaml:"filters"`
	Limit        int             `yaml:"limit"`
	Layout       map[string]ColumnLayout `yaml:"layout"`
}

type PipelineFilters struct {
	DefinitionID int    `yaml:"definitionId"`
	RequestedFor string `yaml:"requestedFor"`
	StatusFilter string `yaml:"statusFilter"`
	ResultFilter string `yaml:"resultFilter"`
	BranchName   string `yaml:"branchName"`
}

type ColumnLayout struct {
	Hidden bool `yaml:"hidden"`
}

type Defaults struct {
	View                   string         `yaml:"view"`
	RefetchIntervalMinutes int            `yaml:"refetchIntervalMinutes"`
	Preview                PreviewConfig  `yaml:"preview"`
	PRsLimit               int            `yaml:"prsLimit"`
	WorkItemsLimit         int            `yaml:"workItemsLimit"`
	PipelinesLimit         int            `yaml:"pipelinesLimit"`
	Layout                 DefaultLayouts `yaml:"layout"`
}

type PreviewConfig struct {
	Open  bool `yaml:"open"`
	Width int  `yaml:"width"`
}

type DefaultLayouts struct {
	PRs       map[string]ColumnLayout `yaml:"prs"`
	WorkItems map[string]ColumnLayout `yaml:"workItems"`
	Pipelines map[string]ColumnLayout `yaml:"pipelines"`
}

type Theme struct {
	Colors ThemeColors `yaml:"colors"`
}

type ThemeColors struct {
	Text     string         `yaml:"text"`
	FaintText string        `yaml:"faintText"`
	Border   string         `yaml:"border"`
	Selected string         `yaml:"selected"`
	Title    string         `yaml:"title"`
	PR       StatusColors   `yaml:"pr"`
	WorkItem StatusColors   `yaml:"workItem"`
	Pipeline StatusColors   `yaml:"pipeline"`
}

type StatusColors struct {
	Open      string `yaml:"open"`
	Draft     string `yaml:"draft"`
	Active    string `yaml:"active"`
	Completed string `yaml:"completed"`
	Abandoned string `yaml:"abandoned"`
	New       string `yaml:"new"`
	Resolved  string `yaml:"resolved"`
	Closed    string `yaml:"closed"`
	Succeeded string `yaml:"succeeded"`
	Failed    string `yaml:"failed"`
	Running   string `yaml:"running"`
	Canceled  string `yaml:"canceled"`
}

type KeybindingsConfig struct {
	Universal []Keybinding `yaml:"universal"`
	PRs       []Keybinding `yaml:"prs"`
	WorkItems []Keybinding `yaml:"workItems"`
	Pipelines []Keybinding `yaml:"pipelines"`
}

type Keybinding struct {
	Key     string `yaml:"key"`
	Builtin string `yaml:"builtin"`
	Command string `yaml:"command"`
	Name    string `yaml:"name"`
}
