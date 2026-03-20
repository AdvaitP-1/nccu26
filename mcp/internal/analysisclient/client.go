// Package analysisclient is a typed HTTP client for the structural analysis
// backend.  It is the only code in the MCP that speaks to /backend.
package analysisclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// Client wraps HTTP calls to the analysis backend.
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

// New creates a Client pointing at *baseURL* with the given timeout.
func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: slog.Default().With("component", "analysisclient"),
	}
}

// AnalyzeOverlaps sends changesets to POST /analyze/overlaps and returns
// the parsed response.
func (c *Client) AnalyzeOverlaps(
	ctx context.Context,
	changesets []models.AnalysisChangeSet,
) (*models.AnalyzeOverlapsResponse, error) {

	reqBody := models.AnalyzeOverlapsRequest{Changesets: changesets}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/analyze/overlaps"
	c.logger.Debug("calling backend", "url", url, "changesets", len(changesets))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("backend request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(body))
	}

	var result models.AnalyzeOverlapsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Info("analysis complete",
		"overlaps", len(result.Overlaps),
		"file_risks", len(result.FileRisks),
	)
	return &result, nil
}

// Health pings GET /health on the backend and returns an error if it fails.
func (c *Client) Health(ctx context.Context) error {
	url := c.baseURL + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("backend health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backend health returned %d", resp.StatusCode)
	}
	return nil
}
