package mixpanel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

func NewLoopbackOAuth() (auth.OAuthHandler, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, fmt.Errorf("listen for OAuth callback: %w", err)
	}

	redirectURL := "http://" + listener.Addr().String() + "/callback"
	log.Printf("OAuth redirect URL: %s", redirectURL)

	callbacks := make(chan *auth.AuthorizationResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if query.Get("code") == "" || query.Get("state") == "" {
			http.Error(w, "Missing OAuth code or state", http.StatusBadRequest)
			return
		}

		select {
		case callbacks <- &auth.AuthorizationResult{
			Code:  query.Get("code"),
			State: query.Get("state"),
			Iss:   query.Get("iss"),
		}:
			_, _ = w.Write([]byte("Callback received. You can close this tab."))
		default:
			http.Error(w, "Callback already received", http.StatusConflict)
		}
	})

	callbackServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	oauthHandler, err := auth.NewAuthorizationCodeHandler(&auth.AuthorizationCodeHandlerConfig{
		RedirectURL: redirectURL,
		DynamicClientRegistrationConfig: &auth.DynamicClientRegistrationConfig{
			Metadata: &oauthex.ClientRegistrationMetadata{
				ClientName:   "App Intelligence Mixpanel Probe",
				RedirectURIs: []string{redirectURL},
			},
		},
		AuthorizationCodeFetcher: func(ctx context.Context, args *auth.AuthorizationArgs) (*auth.AuthorizationResult, error) {
			log.Printf("Open this URL in your browser:\n%s", args.URL)

			select {
			case result := <-callbacks:
				return result, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	})
	if err != nil {
		_ = listener.Close()
		return nil, nil, fmt.Errorf("create OAuth handler: %w", err)
	}

	go func() {
		if err := callbackServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("serve OAuth callback: %v", err)
		}
	}()

	return oauthHandler, func() {
		_ = callbackServer.Close()
		_ = listener.Close()
	}, nil
}
