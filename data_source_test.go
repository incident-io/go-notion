package notion_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dstotijn/go-notion"
	"github.com/google/go-cmp/cmp"
)

func TestFindDatabaseByIDV2(t *testing.T) {
	t.Parallel()

	const (
		dbID       = "00000000-0000-0000-0000-000000000000"
		expPath    = "/v1/databases/" + dbID
		expVersion = "2025-09-03"
	)

	tests := []struct {
		name           string
		respBody       string
		respStatusCode int
		expDatabase    notion.Database
		expError       error
	}{
		{
			name: "successful response with data sources",
			respBody: `{
				"object": "database",
				"id": "668d797c-76fa-4934-9b05-ad288df2d136",
				"created_time": "2020-03-17T19:10:04.968Z",
				"last_edited_time": "2020-03-17T21:49:37.913Z",
				"url": "https://www.notion.so/668d797c76fa49349b05ad288df2d136",
				"title": [],
				"description": [],
				"properties": {},
				"parent": { "type": "page_id", "page_id": "48f8fee9-cd79-4180-bc2f-ec0398253067" },
				"archived": false,
				"is_inline": false,
				"data_sources": [
					{ "id": "d1", "name": "Primary" },
					{ "id": "d2", "name": "Synced Jira" }
				]
			}`,
			respStatusCode: http.StatusOK,
			expDatabase: notion.Database{
				ID:             "668d797c-76fa-4934-9b05-ad288df2d136",
				CreatedTime:    mustParseTime(time.RFC3339Nano, "2020-03-17T19:10:04.968Z"),
				LastEditedTime: mustParseTime(time.RFC3339Nano, "2020-03-17T21:49:37.913Z"),
				URL:            "https://www.notion.so/668d797c76fa49349b05ad288df2d136",
				Title:          []notion.RichText{},
				Description:    []notion.RichText{},
				Properties:     notion.DatabaseProperties{},
				Parent: notion.Parent{
					Type:   notion.ParentTypePage,
					PageID: "48f8fee9-cd79-4180-bc2f-ec0398253067",
				},
				DataSources: []notion.DataSourceReference{
					{ID: "d1", Name: "Primary"},
					{ID: "d2", Name: "Synced Jira"},
				},
			},
		},
		{
			name:           "error response",
			respBody:       `{"object":"error","status":404,"code":"object_not_found","message":"Not found."}`,
			respStatusCode: http.StatusNotFound,
			expError:       errors.New("notion: failed to find database: Not found. (code: object_not_found, status: 404)"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &http.Client{
				Transport: &mockRoundtripper{fn: func(r *http.Request) (*http.Response, error) {
					checkPathAndVersion(t, r, expPath, expVersion)
					return &http.Response{
						StatusCode: tt.respStatusCode,
						Status:     http.StatusText(tt.respStatusCode),
						Body:       ioutil.NopCloser(strings.NewReader(tt.respBody)),
					}, nil
				}},
			}
			client := notion.NewClient("secret-api-key", notion.WithHTTPClient(httpClient))
			db, err := client.FindDatabaseByIDV2(context.Background(), dbID)

			checkError(t, tt.expError, err)

			if diff := cmp.Diff(tt.expDatabase, db); diff != "" {
				t.Fatalf("database not equal (-exp, +got):\n%v", diff)
			}
		})
	}
}

func TestQueryDataSource(t *testing.T) {
	t.Parallel()

	const (
		dsID       = "d1111111-1111-1111-1111-111111111111"
		expPath    = "/v1/data_sources/" + dsID + "/query"
		expVersion = "2025-09-03"
	)

	filterAfter := mustParseTime(time.RFC3339Nano, "2026-01-01T00:00:00.000Z")

	tests := []struct {
		name           string
		query          *notion.DataSourceQuery
		respBody       string
		respStatusCode int
		expPostBody    map[string]interface{}
		expResponse    notion.DataSourceQueryResponse
		expError       error
	}{
		{
			name:           "no query, successful response",
			respBody:       `{"object":"list","results":[],"has_more":false,"next_cursor":null}`,
			respStatusCode: http.StatusOK,
			expResponse: notion.DataSourceQueryResponse{
				Results: []notion.Page{},
			},
		},
		{
			name: "query with last_edited_time filter + data-source-specific fields",
			query: &notion.DataSourceQuery{
				Filter: &notion.DatabaseQueryFilter{
					Timestamp: notion.TimestampLastEditedTime,
					DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
						LastEditedTime: &notion.DatePropertyFilter{After: &filterAfter},
					},
				},
				PageSize:         50,
				FilterProperties: []string{"title", "status"},
				ResultType:       "page",
			},
			respBody:       `{"object":"list","results":[],"has_more":false,"next_cursor":null}`,
			respStatusCode: http.StatusOK,
			expPostBody: map[string]interface{}{
				"filter": map[string]interface{}{
					"timestamp": "last_edited_time",
					"last_edited_time": map[string]interface{}{
						"after": "2026-01-01T00:00:00Z",
					},
				},
				"page_size":         float64(50),
				"filter_properties": []interface{}{"title", "status"},
				"result_type":       "page",
			},
			expResponse: notion.DataSourceQueryResponse{
				Results: []notion.Page{},
			},
		},
		{
			name:           "error response",
			respBody:       `{"object":"error","status":400,"code":"validation_error","message":"bad filter"}`,
			respStatusCode: http.StatusBadRequest,
			expError:       errors.New("notion: failed to query data source: bad filter (code: validation_error, status: 400)"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &http.Client{
				Transport: &mockRoundtripper{fn: func(r *http.Request) (*http.Response, error) {
					checkPathAndVersion(t, r, expPath, expVersion)
					checkPostBody(t, r, tt.expPostBody)
					return &http.Response{
						StatusCode: tt.respStatusCode,
						Status:     http.StatusText(tt.respStatusCode),
						Body:       ioutil.NopCloser(strings.NewReader(tt.respBody)),
					}, nil
				}},
			}
			client := notion.NewClient("secret-api-key", notion.WithHTTPClient(httpClient))
			resp, err := client.QueryDataSource(context.Background(), dsID, tt.query)

			checkError(t, tt.expError, err)

			if diff := cmp.Diff(tt.expResponse, resp); diff != "" {
				t.Fatalf("response not equal (-exp, +got):\n%v", diff)
			}
		})
	}
}

func checkPathAndVersion(t *testing.T, r *http.Request, expPath, expVersion string) {
	t.Helper()
	if got := r.URL.Path; got != expPath {
		t.Errorf("path mismatch (expected: %v, got: %v)", expPath, got)
	}
	if got := r.Header.Get("Notion-Version"); got != expVersion {
		t.Errorf("Notion-Version mismatch (expected: %v, got: %v)", expVersion, got)
	}
}

func checkPostBody(t *testing.T, r *http.Request, exp map[string]interface{}) {
	t.Helper()
	got := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&got)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if len(exp) == 0 && len(got) != 0 {
		t.Errorf("unexpected post body: %+v", got)
		return
	}
	if len(exp) != 0 && len(got) == 0 {
		t.Errorf("post body missing (expected %+v)", exp)
		return
	}
	if len(exp) == 0 {
		return
	}
	if diff := cmp.Diff(exp, got); diff != "" {
		t.Errorf("post body not equal (-exp, +got):\n%v", diff)
	}
}
