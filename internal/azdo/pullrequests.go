package azdo

import (
	"fmt"
	"net/url"
	"strings"
)

type PRSearchCriteria struct {
	Status     string
	CreatorID  string
	ReviewerID string
	Repository string
	TargetRef  string
	SourceRef  string
}

// ListPullRequests fetches PRs at the project level (cross-repo).
// If criteria.Repository is set, scopes to that single repo.
func (c *Client) ListPullRequests(org, project string, criteria PRSearchCriteria, top int) ([]PullRequest, error) {
	var apiURL string
	if criteria.Repository != "" {
		apiURL = fmt.Sprintf("%s/git/repositories/%s/pullrequests", c.projectURL(org, project), criteria.Repository)
	} else {
		apiURL = fmt.Sprintf("%s/git/pullrequests", c.projectURL(org, project))
	}

	params := url.Values{}
	if criteria.Status != "" && criteria.Status != "all" {
		params.Set("searchCriteria.status", criteria.Status)
	}
	if criteria.CreatorID != "" {
		params.Set("searchCriteria.creatorId", c.ResolveMe(criteria.CreatorID))
	}
	if criteria.ReviewerID != "" {
		params.Set("searchCriteria.reviewerId", c.ResolveMe(criteria.ReviewerID))
	}
	if criteria.TargetRef != "" {
		params.Set("searchCriteria.targetRefName", criteria.TargetRef)
	}
	if criteria.SourceRef != "" {
		params.Set("searchCriteria.sourceRefName", criteria.SourceRef)
	}
	if top > 0 {
		params.Set("$top", fmt.Sprintf("%d", top))
	}

	var resp ListResponse[PullRequest]
	if err := c.get(apiURL, params, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// GetPullRequest fetches a single PR with full details.
func (c *Client) GetPullRequest(org, project, repoID string, prID int) (*PullRequest, error) {
	apiURL := fmt.Sprintf("%s/git/repositories/%s/pullrequests/%d", c.projectURL(org, project), repoID, prID)

	var pr PullRequest
	if err := c.get(apiURL, nil, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// VotePullRequest sets a reviewer's vote on a PR.
// vote: 10=approve, 5=approve with suggestions, -5=wait, -10=reject, 0=reset
func (c *Client) VotePullRequest(org, project, repoID string, prID int, reviewerID string, vote int) error {
	apiURL := fmt.Sprintf("%s/git/repositories/%s/pullrequests/%d/reviewers/%s",
		c.projectURL(org, project), repoID, prID, reviewerID)

	body := map[string]interface{}{
		"vote": vote,
	}
	return c.put(apiURL, body, nil)
}

// UpdatePullRequestStatus changes a PR's status (active, completed, abandoned).
func (c *Client) UpdatePullRequestStatus(org, project, repoID string, prID int, status string) error {
	apiURL := fmt.Sprintf("%s/git/repositories/%s/pullrequests/%d",
		c.projectURL(org, project), repoID, prID)

	body := map[string]interface{}{
		"status": status,
	}
	return c.patch(apiURL, body, nil)
}

// GetPullRequestIterations fetches iterations (push events) for a PR.
func (c *Client) GetPullRequestIterations(org, project, repoID string, prID int) ([]PRIteration, error) {
	apiURL := fmt.Sprintf("%s/git/repositories/%s/pullrequests/%d/iterations",
		c.projectURL(org, project), repoID, prID)

	var resp ListResponse[PRIteration]
	if err := c.get(apiURL, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// GetPullRequestChanges fetches the file changes for a specific iteration.
func (c *Client) GetPullRequestChanges(org, project, repoID string, prID, iterationID int) (*PRChanges, error) {
	apiURL := fmt.Sprintf("%s/git/repositories/%s/pullrequests/%d/iterations/%d/changes",
		c.projectURL(org, project), repoID, prID, iterationID)

	var changes PRChanges
	if err := c.get(apiURL, nil, &changes); err != nil {
		return nil, err
	}
	return &changes, nil
}

// GetPullRequestDiff fetches the diff content between iterations for generating unified diffs.
func (c *Client) GetPullRequestDiff(org, project, repoID string, prID int) (string, error) {
	iterations, err := c.GetPullRequestIterations(org, project, repoID, prID)
	if err != nil {
		return "", err
	}
	if len(iterations) == 0 {
		return "", fmt.Errorf("no iterations found for PR #%d", prID)
	}

	lastIter := iterations[len(iterations)-1]
	changes, err := c.GetPullRequestChanges(org, project, repoID, prID, lastIter.ID)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, entry := range changes.ChangeEntries {
		changeType := entry.ChangeType
		path := entry.Item.Path
		sb.WriteString(fmt.Sprintf("--- a%s\n+++ b%s\n", path, path))
		sb.WriteString(fmt.Sprintf("@@ %s @@\n", changeType))
	}
	return sb.String(), nil
}
