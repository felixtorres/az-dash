package azdo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

const azdoResourceID = "499b84ac-1321-427f-aa17-267ca6975798"

type AuthProvider interface {
	GetToken() (string, error)
	AuthHeader() (string, string, error) // header name, header value, error
}

// AzCliAuth uses `az account get-access-token` with caching and auto-refresh.
type AzCliAuth struct {
	mu          sync.Mutex
	cachedToken string
	expiresOn   time.Time
}

type azTokenResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresOn   string `json:"expiresOn"`
}

func NewAzCliAuth() *AzCliAuth {
	return &AzCliAuth{}
}

func (a *AzCliAuth) GetToken() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cachedToken != "" && time.Now().Add(5*time.Minute).Before(a.expiresOn) {
		return a.cachedToken, nil
	}

	return a.refresh()
}

func (a *AzCliAuth) refresh() (string, error) {
	out, err := exec.Command("az", "account", "get-access-token",
		"--resource", azdoResourceID,
		"--output", "json",
	).Output()
	if err != nil {
		return "", fmt.Errorf("az CLI auth failed (is `az` installed and logged in?): %w", err)
	}

	var resp azTokenResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("parsing az token response: %w", err)
	}

	expiry, err := time.Parse("2006-01-02 15:04:05.999999", resp.ExpiresOn)
	if err != nil {
		expiry = time.Now().Add(1 * time.Hour)
	}

	a.cachedToken = resp.AccessToken
	a.expiresOn = expiry
	return a.cachedToken, nil
}

func (a *AzCliAuth) ForceRefresh() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.refresh()
}

func (a *AzCliAuth) AuthHeader() (string, string, error) {
	token, err := a.GetToken()
	if err != nil {
		return "", "", err
	}
	return "Authorization", "Bearer " + token, nil
}

// PatAuth uses a Personal Access Token (from config or env).
type PatAuth struct {
	token string
}

func NewPatAuth(token string) *PatAuth {
	if token == "" {
		token = os.Getenv("AZ_DEVOPS_PAT")
	}
	return &PatAuth{token: token}
}

func (p *PatAuth) GetToken() (string, error) {
	if p.token == "" {
		return "", fmt.Errorf("PAT not configured: set auth.pat in config or AZ_DEVOPS_PAT env var")
	}
	return p.token, nil
}

func (p *PatAuth) AuthHeader() (string, string, error) {
	token, err := p.GetToken()
	if err != nil {
		return "", "", err
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(":" + token))
	return "Authorization", "Basic " + encoded, nil
}
