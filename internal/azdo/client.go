package azdo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiVersion = "7.1"

type Client struct {
	baseURL      string
	organization string
	project      string
	auth         AuthProvider
	httpClient   *http.Client
	userID       string // resolved @me GUID
}

func NewClient(baseURL, organization, project string, auth AuthProvider) *Client {
	return &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		organization: organization,
		project:      project,
		auth:         auth,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Bootstrap resolves the current user's ID for @me substitution.
func (c *Client) Bootstrap() error {
	profile, err := c.getProfile()
	if err != nil {
		return fmt.Errorf("resolving user profile: %w", err)
	}
	c.userID = profile.ID
	return nil
}

func (c *Client) UserID() string {
	return c.userID
}

// ResolveMe replaces "@me" with the actual user GUID.
func (c *Client) ResolveMe(value string) string {
	if value == "@me" {
		return c.userID
	}
	return value
}

func (c *Client) getProfile() (*Profile, error) {
	req, err := http.NewRequest("GET", "https://app.vssps.visualstudio.com/_apis/profile/profiles/me?api-version="+apiVersion, nil)
	if err != nil {
		return nil, err
	}

	headerName, headerValue, err := c.auth.AuthHeader()
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerName, headerValue)

	var profile Profile
	if err := c.doRequest(req, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// projectURL builds the base URL for project-scoped API calls.
func (c *Client) projectURL(org, project string) string {
	if org == "" {
		org = c.organization
	}
	if project == "" {
		project = c.project
	}
	return fmt.Sprintf("%s/%s/%s/_apis", c.baseURL, org, project)
}

// orgURL builds the base URL for org-level API calls (no project).
func (c *Client) orgURL(org string) string {
	if org == "" {
		org = c.organization
	}
	return fmt.Sprintf("%s/%s/_apis", c.baseURL, org)
}

func (c *Client) get(apiURL string, params url.Values, result interface{}) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("api-version", apiVersion)

	fullURL := apiURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}

	return c.doWithAuth(req, result)
}

func (c *Client) post(apiURL string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	fullURL := apiURL + "?api-version=" + apiVersion
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doWithAuth(req, result)
}

func (c *Client) doWithAuth(req *http.Request, result interface{}) error {
	headerName, headerValue, err := c.auth.AuthHeader()
	if err != nil {
		return err
	}
	req.Header.Set(headerName, headerValue)

	err = c.doRequest(req, result)
	if err != nil && isUnauthorized(err) {
		if azAuth, ok := c.auth.(*AzCliAuth); ok {
			if _, refreshErr := azAuth.ForceRefresh(); refreshErr != nil {
				return err
			}
			headerName, headerValue, _ = c.auth.AuthHeader()
			req.Header.Set(headerName, headerValue)
			return c.doRequest(req, result)
		}
	}
	return err
}

func (c *Client) doRequest(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}
	return nil
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Azure DevOps API error %d: %s", e.StatusCode, e.Message)
}

func isUnauthorized(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 401
	}
	return false
}
