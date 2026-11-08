package mixpanel

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestSavedOAuthSession(t *testing.T) {
	const endpoint = "https://mcp-eu.mixpanel.com/mcp"
	path := filepath.Join(t.TempDir(), "mixpanel-login.json")
	if _, err := readOAuthSession(path, endpoint); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("read missing login: %v", err)
	}

	session := savedOAuthSession{
		Endpoint: endpoint,
		Token:    oauth2.Token{AccessToken: "test-token", Expiry: time.Now().Add(time.Hour)},
	}
	if err := writeOAuthSession(path, session); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("login file permissions = %o, want 600", info.Mode().Perm())
	}

	handler, err := readOAuthSession(path, endpoint)
	if err != nil {
		t.Fatal(err)
	}
	source, err := handler.TokenSource(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	token, err := source.Token()
	if err != nil || token.AccessToken != session.Token.AccessToken {
		t.Fatalf("read saved access token: %v", err)
	}

	if _, err := readOAuthSession(path, "https://mcp.mixpanel.com/mcp"); err == nil || !strings.Contains(err.Error(), "different endpoint") {
		t.Fatalf("read login for different endpoint: %v", err)
	}

	session.Token.Expiry = time.Now().Add(-time.Hour)
	if err := writeOAuthSession(path, session); err != nil {
		t.Fatal(err)
	}
	if _, err := readOAuthSession(path, endpoint); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("read expired login: %v", err)
	}
}
