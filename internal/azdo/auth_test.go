package azdo

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"
)

func TestPatAuthHeader(t *testing.T) {
	auth := NewPatAuth("my-secret-pat")
	name, value, err := auth.AuthHeader()
	if err != nil {
		t.Fatal(err)
	}
	if name != "Authorization" {
		t.Errorf("header name = %q, want Authorization", name)
	}
	if !strings.HasPrefix(value, "Basic ") {
		t.Errorf("header value should start with 'Basic ', got %q", value)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, "Basic "))
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != ":my-secret-pat" {
		t.Errorf("decoded = %q, want %q", string(decoded), ":my-secret-pat")
	}
}

func TestPatAuthFromEnv(t *testing.T) {
	os.Setenv("AZ_DEVOPS_PAT", "env-token")
	defer os.Unsetenv("AZ_DEVOPS_PAT")

	auth := NewPatAuth("")
	token, err := auth.GetToken()
	if err != nil {
		t.Fatal(err)
	}
	if token != "env-token" {
		t.Errorf("token = %q, want %q", token, "env-token")
	}
}

func TestPatAuthEmpty(t *testing.T) {
	os.Unsetenv("AZ_DEVOPS_PAT")
	auth := NewPatAuth("")
	_, err := auth.GetToken()
	if err == nil {
		t.Fatal("expected error for empty PAT")
	}
}

func TestResolveMe(t *testing.T) {
	client := &Client{userID: "abc-123"}

	if got := client.ResolveMe("@me"); got != "abc-123" {
		t.Errorf("ResolveMe(@me) = %q, want %q", got, "abc-123")
	}
	if got := client.ResolveMe("other-id"); got != "other-id" {
		t.Errorf("ResolveMe(other-id) = %q, want %q", got, "other-id")
	}
}
