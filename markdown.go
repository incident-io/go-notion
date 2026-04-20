package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Optional query params.
type PageMarkdownOpts struct {
	IncludeTranscript bool
}

// See: https://developers.notion.com/reference/retrieve-page-markdown
type PageMarkdownResponse struct {
	Markdown        string   `json:"markdown"`
	Truncated       bool     `json:"truncated"`
	UnknownBlockIDs []string `json:"unknown_block_ids"`
}

// Public integrations only. Pins Notion-Version 2026-03-11.
// See: https://developers.notion.com/reference/retrieve-page-markdown
func (c *Client) FindPageMarkdownByID(ctx context.Context, pageID string, opts *PageMarkdownOpts) (result PageMarkdownResponse, err error) {
	path := "/pages/" + pageID + "/markdown"
	if opts != nil && opts.IncludeTranscript {
		path += "?include_transcript=true"
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return PageMarkdownResponse{}, fmt.Errorf("notion: invalid request: %w", err)
	}
	req.Header.Set("Notion-Version", "2026-03-11")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return PageMarkdownResponse{}, fmt.Errorf("notion: failed to make HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return PageMarkdownResponse{}, fmt.Errorf("notion: failed to find page markdown: %w", parseErrorResponse(res))
	}

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return PageMarkdownResponse{}, fmt.Errorf("notion: failed to parse HTTP response: %w", err)
	}

	return result, nil
}
