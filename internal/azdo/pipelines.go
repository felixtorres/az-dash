package azdo

import (
	"fmt"
	"io"
	"net/http"
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

// CancelBuild sets a build's status to cancelling.
func (c *Client) CancelBuild(org, project string, buildID int) error {
	apiURL := fmt.Sprintf("%s/build/builds/%d", c.projectURL(org, project), buildID)
	body := map[string]interface{}{
		"status": "cancelling",
	}
	return c.patch(apiURL, body, nil)
}

// GetBuildLogs fetches all log lines for a build.
func (c *Client) GetBuildLogs(org, project string, buildID int) ([]BuildLog, error) {
	apiURL := fmt.Sprintf("%s/build/builds/%d/logs", c.projectURL(org, project), buildID)

	var resp ListResponse[BuildLog]
	if err := c.get(apiURL, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// GetBuildLogContent fetches the text content of a specific log.
func (c *Client) GetBuildLogContent(org, project string, buildID, logID int) (string, error) {
	apiURL := fmt.Sprintf("%s/build/builds/%d/logs/%d", c.projectURL(org, project), buildID, logID)
	params := url.Values{}
	params.Set("api-version", apiVersion)

	fullURL := apiURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return "", err
	}

	headerName, headerValue, err := c.auth.AuthHeader()
	if err != nil {
		return "", err
	}
	req.Header.Set(headerName, headerValue)
	req.Header.Set("Accept", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	return string(body), nil
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
