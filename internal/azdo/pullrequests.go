package azdo

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
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

func (c *Client) GetBlobContent(org, project, repoID, objectID string) (string, error) {
	apiURL := fmt.Sprintf("%s/git/repositories/%s/blobs/%s", c.projectURL(org, project), repoID, objectID)
	body, err := c.getRaw(apiURL, nil, "text/plain")
	if err != nil {
		return "", err
	}
	return string(body), nil
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
		diff, err := c.buildChangeDiff(org, project, repoID, entry)
		if err != nil {
			sb.WriteString(fmt.Sprintf("--- a%s\n+++ b%s\n", entryPathOld(entry), entryPathNew(entry)))
			sb.WriteString(fmt.Sprintf("@@ %s @@\n", entry.ChangeType))
			sb.WriteString(fmt.Sprintf("[unable to load file contents: %v]\n", err))
			continue
		}
		sb.WriteString(diff)
	}
	return sb.String(), nil
}

func (c *Client) buildChangeDiff(org, project, repoID string, entry PRChangeEntry) (string, error) {
	oldPath := entryPathOld(entry)
	newPath := entryPathNew(entry)
	oldContent, err := c.contentForChange(org, project, repoID, entry, true)
	if err != nil {
		return "", err
	}
	newContent, err := c.contentForChange(org, project, repoID, entry, false)
	if err != nil {
		return "", err
	}

	ud := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldContent),
		B:        difflib.SplitLines(newContent),
		FromFile: "a" + oldPath,
		ToFile:   "b" + newPath,
		Context:  3,
	}

	text, err := difflib.GetUnifiedDiffString(ud)
	if err != nil {
		return "", err
	}
	if text == "" {
		return fmt.Sprintf("--- a%s\n+++ b%s\n", oldPath, newPath), nil
	}
	return text, nil
}

func (c *Client) contentForChange(org, project, repoID string, entry PRChangeEntry, old bool) (string, error) {
	changeType := normalizeChangeType(entry.ChangeType)
	if old {
		if changeType == "add" {
			return "", nil
		}
		path := entryPathOld(entry)
		if path == "" {
			return "", nil
		}
		version := entry.Item.OriginalObjectID
		if version == "" {
			version = entry.Item.ObjectID
		}
		if version == "" {
			return "", nil
		}
		return c.GetBlobContent(org, project, repoID, version)
	}

	if changeType == "delete" {
		return "", nil
	}
	path := entryPathNew(entry)
	if path == "" {
		return "", nil
	}
	if entry.NewContent != nil && entry.NewContent.ContentType == "rawText" {
		return entry.NewContent.Content, nil
	}
	version := entry.Item.ObjectID
	if version == "" {
		return "", nil
	}
	return c.GetBlobContent(org, project, repoID, version)
}

func entryPathOld(entry PRChangeEntry) string {
	if entry.OriginalPath != "" {
		return entry.OriginalPath
	}
	return entry.Item.Path
}

func entryPathNew(entry PRChangeEntry) string {
	return entry.Item.Path
}

func normalizeChangeType(changeType string) string {
	parts := strings.Split(changeType, ",")
	if len(parts) == 0 {
		return changeType
	}
	return strings.TrimSpace(parts[0])
}
