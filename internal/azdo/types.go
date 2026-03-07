package azdo

import (
	"fmt"
	"time"
)

// Profile represents the current user's Azure DevOps profile.
type Profile struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// PullRequest represents an Azure DevOps pull request.
type PullRequest struct {
	PullRequestID int           `json:"pullRequestId"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	Status        string        `json:"status"` // active, completed, abandoned
	CreatedBy     IdentityRef   `json:"createdBy"`
	CreationDate  time.Time     `json:"creationDate"`
	ClosedDate    *time.Time    `json:"closedDate"`
	SourceRefName string        `json:"sourceRefName"`
	TargetRefName string        `json:"targetRefName"`
	MergeStatus   string        `json:"mergeStatus"`
	IsDraft       bool          `json:"isDraft"`
	Repository    GitRepository `json:"repository"`
	Reviewers     []Reviewer    `json:"reviewers"`
	Labels        []Label       `json:"labels"`
	URL           string        `json:"url"`
}

func (pr *PullRequest) WebURL() string {
	if pr.Repository.WebURL != "" {
		return pr.Repository.WebURL + "/pullrequest/" + fmt.Sprintf("%d", pr.PullRequestID)
	}
	return ""
}

type IdentityRef struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
}

type Reviewer struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	Vote        int    `json:"vote"` // 10=approved, 5=approved with suggestions, -5=waiting, -10=rejected, 0=no response
	IsRequired  bool   `json:"isRequired"`
}

type Label struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GitRepository struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	WebURL string `json:"webUrl"`
}

// WorkItem represents an Azure DevOps work item.
type WorkItem struct {
	ID     int                    `json:"id"`
	Rev    int                    `json:"rev"`
	Fields map[string]interface{} `json:"fields"`
	URL    string                 `json:"url"`
}

func (wi *WorkItem) Field(name string) interface{} {
	return wi.Fields[name]
}

func (wi *WorkItem) StringField(name string) string {
	v, ok := wi.Fields[name]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case map[string]interface{}:
		if dn, ok := val["displayName"].(string); ok {
			return dn
		}
	}
	return fmt.Sprintf("%v", v)
}

func (wi *WorkItem) WebURL() string {
	// Work items don't have a direct webUrl in the response; construct from org info.
	// The `url` field points to the API URL, which we can't easily convert.
	// Callers should construct the URL using org/project context.
	return wi.URL
}

// PRIteration represents a push event on a pull request.
type PRIteration struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	CreatedDate time.Time `json:"createdDate"`
}

// PRChanges represents the file changes in a PR iteration.
type PRChanges struct {
	ChangeEntries []PRChangeEntry `json:"changeEntries"`
}

type PRChangeEntry struct {
	ChangeType   string       `json:"changeType"` // add, edit, delete, rename
	Item         PRChangeItem `json:"item"`
	OriginalPath string       `json:"originalPath"`
	NewContent   *ItemContent `json:"newContent"`
}

type PRChangeItem struct {
	Path             string `json:"path"`
	ObjectID         string `json:"objectId"`
	OriginalObjectID string `json:"originalObjectId"`
}

type ItemContent struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

// Build represents an Azure DevOps build (pipeline run).
type Build struct {
	ID            int             `json:"id"`
	BuildNumber   string          `json:"buildNumber"`
	Status        string          `json:"status"` // notStarted, inProgress, completed, cancelling
	Result        string          `json:"result"` // succeeded, failed, canceled, partiallySucceeded
	QueueTime     time.Time       `json:"queueTime"`
	StartTime     *time.Time      `json:"startTime"`
	FinishTime    *time.Time      `json:"finishTime"`
	SourceBranch  string          `json:"sourceBranch"`
	SourceVersion string          `json:"sourceVersion"`
	RequestedFor  IdentityRef     `json:"requestedFor"`
	Definition    BuildDefinition `json:"definition"`
	Project       TeamProject     `json:"project"`
	URL           string          `json:"url"`
}

func (b *Build) WebURL() string {
	if b.Project.Name != "" && b.Definition.ID > 0 {
		return fmt.Sprintf("https://dev.azure.com/%s/_build/results?buildId=%d", b.Project.Name, b.ID)
	}
	return ""
}

func (b *Build) Duration() time.Duration {
	if b.StartTime == nil || b.FinishTime == nil {
		return 0
	}
	return b.FinishTime.Sub(*b.StartTime)
}

type BuildDefinition struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TeamProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Timeline represents build stages/jobs.
type Timeline struct {
	Records []TimelineRecord `json:"records"`
}

type TimelineRecord struct {
	ID         string     `json:"id"`
	ParentID   string     `json:"parentId"`
	Name       string     `json:"name"`
	Type       string     `json:"type"` // Stage, Job, Task
	State      string     `json:"state"`
	Result     string     `json:"result"`
	StartTime  *time.Time `json:"startTime"`
	FinishTime *time.Time `json:"finishTime"`
	Order      int        `json:"order"`
}

// BuildLog represents a log entry for a build.
type BuildLog struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	LineCount int    `json:"lineCount"`
}

// WIQL response.
type WIQLResult struct {
	WorkItems []WIQLWorkItemRef `json:"workItems"`
}

type WIQLWorkItemRef struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

// Generic list response wrapper.
type ListResponse[T any] struct {
	Count int `json:"count"`
	Value []T `json:"value"`
}
