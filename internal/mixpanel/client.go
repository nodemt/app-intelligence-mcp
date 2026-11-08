package mixpanel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	endpoint   string
	username   string
	secret     string
	httpClient *http.Client
}

func NewClient(endpoint, username, secret string) *Client {
	return &Client{
		endpoint:   endpoint,
		username:   username,
		secret:     secret,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) GetSavedReport(ctx context.Context, projectID, bookmarkID int) (SavedReport, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return SavedReport{}, fmt.Errorf("build Mixpanel request %w", err)
	}

	query := req.URL.Query()
	query.Set("project_id", strconv.Itoa(projectID))
	query.Set("bookmark_id", strconv.Itoa(bookmarkID))
	req.URL.RawQuery = query.Encode()

	req.SetBasicAuth(c.username, c.secret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return SavedReport{}, fmt.Errorf("query Mixpanel saved report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SavedReport{}, fmt.Errorf("Mixpanel saved report: HTTP %d", resp.StatusCode)
	}

	var report SavedReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return SavedReport{}, fmt.Errorf("decode Mixpanel saved report: %w", err)
	}

	return report, nil
}
