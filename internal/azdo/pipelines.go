package azdo

import (
	"fmt"
	"net/url"
)

type BuildSearchCriteria struct {
	DefinitionID int
	RequestedFor string
	StatusFilter string
	ResultFilter string
	BranchName   string
}

// ListBuilds fetches builds (pipeline runs) using the Builds API.
func (c *Client) ListBuilds(org, project string, criteria BuildSearchCriteria, top int) ([]Build, error) {
	apiURL := fmt.Sprintf("%s/build/builds", c.projectURL(org, project))

	params := url.Values{}
	if criteria.DefinitionID > 0 {
		params.Set("definitions", fmt.Sprintf("%d", criteria.DefinitionID))
	}
	if criteria.RequestedFor != "" {
		params.Set("requestedFor", c.ResolveMe(criteria.RequestedFor))
	}
	if criteria.StatusFilter != "" {
		params.Set("statusFilter", criteria.StatusFilter)
	}
	if criteria.ResultFilter != "" {
		params.Set("resultFilter", criteria.ResultFilter)
	}
	if criteria.BranchName != "" {
		branch := criteria.BranchName
		if !startsWith(branch, "refs/") {
			branch = "refs/heads/" + branch
		}
		params.Set("branchName", branch)
	}
	if top > 0 {
		params.Set("$top", fmt.Sprintf("%d", top))
	}

	var resp ListResponse[Build]
	if err := c.get(apiURL, params, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// GetBuildTimeline fetches stages/jobs/tasks for a build.
func (c *Client) GetBuildTimeline(org, project string, buildID int) (*Timeline, error) {
	apiURL := fmt.Sprintf("%s/build/builds/%d/timeline", c.projectURL(org, project), buildID)

	var timeline Timeline
	if err := c.get(apiURL, nil, &timeline); err != nil {
		return nil, err
	}
	return &timeline, nil
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
