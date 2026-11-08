package mixpanel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"app-intelligence-mcp/internal/constants"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"golang.org/x/oauth2"
)

type savedOAuthSession struct {
	Endpoint string       `json:"endpoint"`
	Token    oauth2.Token `json:"token"`
}

type savedOAuthHandler struct {
	token oauth2.Token
}

func (h savedOAuthHandler) TokenSource(context.Context) (oauth2.TokenSource, error) {
	if !h.token.Valid() {
		return nil, errors.New("Mixpanel login expired; run mixpanel-login")
	}
	return oauth2.StaticTokenSource(&h.token), nil
}

func (h savedOAuthHandler) Authorize(_ context.Context, _ *http.Request, resp *http.Response) error {
	_ = resp.Body.Close()
	return errors.New("Mixpanel rejected the saved login; run mixpanel-login")
}

func SaveOAuthSession(ctx context.Context, endpoint string, handler auth.OAuthHandler) error {
	source, err := handler.TokenSource(ctx)
	if err != nil {
		return fmt.Errorf("get Mixpanel token source: %w", err)
	}
	if source == nil {
		return errors.New("Mixpanel did not issue an OAuth token")
	}
	token, err := source.Token()
	if err != nil {
		return fmt.Errorf("get Mixpanel access token: %w", err)
	}
	if !token.Valid() {
		return errors.New("Mixpanel did not issue a valid access token")
	}
	path, err := oauthSessionPath()
	if err != nil {
		return err
	}
	return writeOAuthSession(path, savedOAuthSession{
		Endpoint: endpoint,
		Token: oauth2.Token{
			AccessToken: token.AccessToken,
			TokenType:   token.TokenType,
			Expiry:      token.Expiry,
		},
	})
}

func LoadOAuthSession(endpoint string) (auth.OAuthHandler, error) {
	path, err := oauthSessionPath()
	if err != nil {
		return nil, err
	}
	return readOAuthSession(path, endpoint)
}

func oauthSessionPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate Mixpanel login directory: %w", err)
	}
	return filepath.Join(dir, constants.AppName, "mixpanel-login.json"), nil
}

func writeOAuthSession(path string, session savedOAuthSession) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create Mixpanel login directory: %w", err)
	}
	file, err := os.CreateTemp(dir, ".mixpanel-login-*")
	if err != nil {
		return fmt.Errorf("create Mixpanel login file: %w", err)
	}
	defer os.Remove(file.Name())
	err = json.NewEncoder(file).Encode(session)
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("write Mixpanel login: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("close Mixpanel login file: %w", closeErr)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("save Mixpanel login: %w", err)
	}
	return nil
}

func readOAuthSession(path, endpoint string) (auth.OAuthHandler, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("Mixpanel login missing; run mixpanel-login")
	}
	if err != nil {
		return nil, fmt.Errorf("read Mixpanel login: %w", err)
	}
	var session savedOAuthSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, errors.New("Mixpanel login file is invalid; run mixpanel-login")
	}
	if session.Endpoint != endpoint {
		return nil, errors.New("Mixpanel login is for a different endpoint; run mixpanel-login")
	}
	if !session.Token.Valid() {
		return nil, errors.New("Mixpanel login expired; run mixpanel-login")
	}
	return savedOAuthHandler{token: session.Token}, nil
}
