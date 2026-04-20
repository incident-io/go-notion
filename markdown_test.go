package notion_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/dstotijn/go-notion"
	"github.com/google/go-cmp/cmp"
)

func TestFindPageMarkdownByID(t *testing.T) {
	t.Parallel()

	const (
		pageID         = "00000000-0000-0000-0000-000000000000"
		expPath        = "/v1/pages/" + pageID + "/markdown"
		expVersion     = "2026-03-11"
	)

	tests := []struct {
		name           string
		opts           *notion.PageMarkdownOpts
		respBody       string
		respStatusCode int
		expResponse    notion.PageMarkdownResponse
		expError       error
	}{
		{
			name:           "no opts, successful response",
			respBody:       `{"markdown":"# Hello","truncated":false,"unknown_block_ids":[]}`,
			respStatusCode: http.StatusOK,
			expResponse: notion.PageMarkdownResponse{
				Markdown:        "# Hello",
				UnknownBlockIDs: []string{},
			},
		},
		{
			name:           "include_transcript appends query param",
			opts:           &notion.PageMarkdownOpts{IncludeTranscript: true},
			respBody:       `{"markdown":"t","truncated":false,"unknown_block_ids":[]}`,
			respStatusCode: http.StatusOK,
			expResponse: notion.PageMarkdownResponse{
				Markdown:        "t",
				UnknownBlockIDs: []string{},
			},
		},
		{
			name:           "truncated with unknown block IDs",
			respBody:       `{"markdown":"p","truncated":true,"unknown_block_ids":["abc","def"]}`,
			respStatusCode: http.StatusOK,
			expResponse: notion.PageMarkdownResponse{
				Markdown:        "p",
				Truncated:       true,
				UnknownBlockIDs: []string{"abc", "def"},
			},
		},
		{
			name:           "error response",
			respBody:       `{"object":"error","status":404,"code":"object_not_found","message":"Not found."}`,
			respStatusCode: http.StatusNotFound,
			expError:       errors.New("notion: failed to find page markdown: Not found. (code: object_not_found, status: 404)"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &http.Client{
				Transport: &mockRoundtripper{fn: func(r *http.Request) (*http.Response, error) {
					if got := r.URL.Path; got != expPath {
						t.Errorf("path mismatch (expected: %v, got: %v)", expPath, got)
					}
					if got := r.Header.Get("Notion-Version"); got != expVersion {
						t.Errorf("Notion-Version mismatch (expected: %v, got: %v)", expVersion, got)
					}
					if tt.opts != nil && tt.opts.IncludeTranscript {
						if got := r.URL.Query().Get("include_transcript"); got != "true" {
							t.Errorf("include_transcript query param not set (got: %q)", got)
						}
					}
					return &http.Response{
						StatusCode: tt.respStatusCode,
						Status:     http.StatusText(tt.respStatusCode),
						Body:       ioutil.NopCloser(strings.NewReader(tt.respBody)),
					}, nil
				}},
			}
			client := notion.NewClient("secret-api-key", notion.WithHTTPClient(httpClient))
			resp, err := client.FindPageMarkdownByID(context.Background(), pageID, tt.opts)

			checkError(t, tt.expError, err)

			if diff := cmp.Diff(tt.expResponse, resp); diff != "" {
				t.Fatalf("response not equal (-exp, +got):\n%v", diff)
			}
		})
	}
}

func checkError(t *testing.T, exp, got error) {
	t.Helper()
	if exp == nil && got != nil {
		t.Fatalf("unexpected error: %v", got)
	}
	if exp != nil && got == nil {
		t.Fatalf("error not equal (expected: %v, got: nil)", exp)
	}
	if exp != nil && got != nil && exp.Error() != got.Error() {
		t.Fatalf("error not equal (expected: %v, got: %v)", exp, got)
	}
}

