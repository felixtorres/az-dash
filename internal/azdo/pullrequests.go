package azdo

import (
	"fmt"
	"net/url"
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
