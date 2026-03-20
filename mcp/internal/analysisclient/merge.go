// merge.go extends the analysis Client with methods for the backend's
// POST /merge endpoint (merge base content + multiple diffs).

package analysisclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// Merge sends base content and a set of diff payloads to the backend's
// POST /merge endpoint and returns the merge result.
func (c *Client) Merge(
	ctx context.Context,
	baseContent string,
	diffs []string,
) (*models.BackendMergeResponse, error) {

	reqBody := models.BackendMergeRequest{
		BaseContent: baseContent,
		Diffs:       diffs,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal merge request: %w", err)
	}

	url := c.baseURL + "/merge"
	c.logger.Debug("calling backend merge", "url", url, "diffs", len(diffs))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build merge request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("backend merge request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read merge response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend merge returned %d: %s", resp.StatusCode, string(body))
	}

	var result models.BackendMergeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode merge response: %w", err)
	}

	c.logger.Info("merge complete",
		"success", result.Success,
		"conflicts", len(result.Conflicts),
	)
	return &result, nil
}
