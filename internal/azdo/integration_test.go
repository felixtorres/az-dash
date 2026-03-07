package azdo

import (
	"fmt"
	"os"
	"testing"
)

// Integration tests — run with: AZ_DEVOPS_PAT=xxx go test ./internal/azdo/ -run Integration -v
// Skipped when PAT is not set.

const (
	integrationOrg     = "cbre"
	integrationProject = "EnterpriseDataPlatform"
)

func integrationClient(t *testing.T) *Client {
	t.Helper()
	pat := os.Getenv("AZ_DEVOPS_PAT")
	if pat == "" {
		t.Skip("AZ_DEVOPS_PAT not set, skipping integration test")
	}
	auth := NewPatAuth(pat)
	c := NewClient("https://dev.azure.com", integrationOrg, integrationProject, auth)
	return c
}

func TestIntegrationBootstrap(t *testing.T) {
	c := integrationClient(t)
	err := c.Bootstrap()
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	if c.UserID() == "" {
		t.Fatal("UserID empty after bootstrap")
	}
	fmt.Printf("  Resolved user: %s\n", c.UserID())
}

func TestIntegrationListPRs(t *testing.T) {
	c := integrationClient(t)
	if err := c.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	prs, err := c.ListPullRequests("", "", PRSearchCriteria{Status: "active"}, 5)
	if err != nil {
		t.Fatalf("ListPullRequests: %v", err)
	}

	fmt.Printf("  Found %d active PRs\n", len(prs))
	for _, pr := range prs {
		fmt.Printf("    #%d %s [%s] by %s → %s\n",
			pr.PullRequestID, pr.Title, pr.Status,
			pr.CreatedBy.DisplayName, trimRef(pr.TargetRefName))
	}
}

func TestIntegrationListPRsMine(t *testing.T) {
	c := integrationClient(t)
	if err := c.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	prs, err := c.ListPullRequests("", "", PRSearchCriteria{
		CreatorID: "@me",
		Status:    "active",
	}, 10)
	if err != nil {
		t.Fatalf("ListPullRequests @me: %v", err)
	}
	fmt.Printf("  Found %d of my active PRs\n", len(prs))
	for _, pr := range prs {
		fmt.Printf("    #%d %s\n", pr.PullRequestID, pr.Title)
	}
}

func TestIntegrationListBuilds(t *testing.T) {
	c := integrationClient(t)
	if err := c.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	builds, err := c.ListBuilds("", "", BuildSearchCriteria{}, 5)
	if err != nil {
		t.Fatalf("ListBuilds: %v", err)
	}

	fmt.Printf("  Found %d builds\n", len(builds))
	for _, b := range builds {
		result := b.Result
		if result == "" {
			result = b.Status
		}
		fmt.Printf("    #%d %s [%s] %s by %s\n",
			b.ID, b.Definition.Name, result,
			trimRef(b.SourceBranch), b.RequestedFor.DisplayName)
	}
}

func TestIntegrationQueryWorkItems(t *testing.T) {
	c := integrationClient(t)
	if err := c.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	wiql := "SELECT [System.Id] FROM WorkItems WHERE [System.AssignedTo] = @me AND [System.State] <> 'Closed' AND [System.State] <> 'Removed' ORDER BY [System.ChangedDate] DESC"
	items, err := c.QueryWorkItems("", "", wiql, 10)
	if err != nil {
		t.Fatalf("QueryWorkItems: %v", err)
	}

	fmt.Printf("  Found %d work items\n", len(items))
	for _, wi := range items {
		fmt.Printf("    #%d [%s] %s — %s\n",
			wi.ID,
			wi.StringField("System.WorkItemType"),
			wi.StringField("System.Title"),
			wi.StringField("System.State"))
	}
}

func trimRef(ref string) string {
	if len(ref) > 11 && ref[:11] == "refs/heads/" {
		return ref[11:]
	}
	return ref
}
