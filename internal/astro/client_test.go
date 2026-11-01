package astro

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestKeywordPerformanceArgumentsWithAppID(t *testing.T) {
	arguments := keywordPerformanceArguments(
		"photo editor", "es", "iphone", "123456789",
	)

	value, exists := arguments["appId"]
	if !exists {
		t.Fatal("appId is missing")
	}

	if value != "123456789" {
		t.Errorf("appId = %v, want 123456789", value)
	}

	if _, exists := arguments["appID"]; exists {
		t.Error("unexpected appID key; Astro expects appId")
	}
}

func TestKeywordPerformanceArgumentsWithoutAppID(t *testing.T) {
	arguments := keywordPerformanceArguments(
		"photo editor", "es", "iphone", "",
	)

	if _, exists := arguments["appId"]; exists {
		t.Error("appId must be omitted when empty")
	}
}

func TestConnectRespectsDeadline(t *testing.T) {
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var message struct {
				ID json.RawMessage `json:"id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}

			// Accept notifications so cancellation cleanup can finish.
			if len(message.ID) == 0 {
				w.WriteHeader(http.StatusAccepted)
				return
			}

			// Leave requests unanswered until canceled or released.
			select {
			case <-r.Context().Done():
			case <-release:
			}
		},
	))
	defer server.Close()
	defer close(release)

	testCtx, cancel := context.WithTimeout(
		context.Background(),
		100*time.Millisecond,
	)
	defer cancel()

	client, err := Connect(testCtx, server.URL)
	if client != nil {
		defer client.Close()
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got error %v, want context deadline exceeded", err)
	}
}
