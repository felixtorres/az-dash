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
	baseURL         string
	organization    string
	project         string
	auth            AuthProvider
	httpClient      *http.Client
	userID          string // resolved @me GUID
	userDisplayName string
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
// Uses the org-scoped connectionData endpoint which works with both
// PATs and az CLI tokens (the VSSPS profile endpoint requires broader scoping).
func (c *Client) Bootstrap() error {
	connData, err := c.getConnectionData()
	if err != nil {
		return fmt.Errorf("resolving user identity: %w", err)
	}
	c.userID = connData.AuthenticatedUser.ID
	c.userDisplayName = connData.AuthenticatedUser.ProviderDisplayName
	return nil
}

func (c *Client) UserID() string {
	return c.userID
}

func (c *Client) UserDisplayName() string {
	return c.userDisplayName
}

// ResolveMe replaces "@me" with the actual user GUID.
func (c *Client) ResolveMe(value string) string {
	if value == "@me" {
		return c.userID
	}
	return value
}

type connectionData struct {
	AuthenticatedUser connectionUser `json:"authenticatedUser"`
}

type connectionUser struct {
	ID                  string `json:"id"`
	ProviderDisplayName string `json:"providerDisplayName"`
}

func (c *Client) getConnectionData() (*connectionData, error) {
	apiURL := fmt.Sprintf("%s/%s/_apis/connectionData", c.baseURL, c.organization)
	params := url.Values{}
	params.Set("api-version", apiVersion+"-preview")

	fullURL := apiURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	headerName, headerValue, err := c.auth.AuthHeader()
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerName, headerValue)

	var data connectionData
	if err := c.doRequest(req, &data); err != nil {
		return nil, err
	}
	return &data, nil
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

// postRaw sends a POST without appending api-version (caller handles it).
func (c *Client) postRaw(fullURL string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", fullURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doWithAuth(req, result)
}

// patchJSON sends a PATCH with application/json-patch+json content type
// (required for Azure DevOps work item updates).
func (c *Client) patchJSON(apiURL string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	fullURL := apiURL + "?api-version=" + apiVersion
	req, err := http.NewRequest("PATCH", fullURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json-patch+json")

	return c.doWithAuth(req, result)
}

func (c *Client) put(apiURL string, body interface{}, result interface{}) error {
	return c.mutate("PUT", apiURL, body, result)
}

func (c *Client) patch(apiURL string, body interface{}, result interface{}) error {
	return c.mutate("PATCH", apiURL, body, result)
}

func (c *Client) mutate(method, apiURL string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	fullURL := apiURL + "?api-version=" + apiVersion
	req, err := http.NewRequest(method, fullURL, strings.NewReader(string(jsonBody)))
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
