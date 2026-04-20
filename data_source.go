package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Populates Database.DataSources.
type DataSourceReference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// See: https://developers.notion.com/reference/query-a-data-source
type DataSourceQuery struct {
	Filter           *DatabaseQueryFilter `json:"filter,omitempty"`
	Sorts            []DatabaseQuerySort  `json:"sorts,omitempty"`
	StartCursor      string               `json:"start_cursor,omitempty"`
	PageSize         int                  `json:"page_size,omitempty"`
	FilterProperties []string             `json:"filter_properties,omitempty"`
	InTrash          bool                 `json:"in_trash,omitempty"`
	ResultType       string               `json:"result_type,omitempty"` // page | data_source
}

// See: https://developers.notion.com/reference/query-a-data-source
type DataSourceQueryResponse struct {
	Results    []Page  `json:"results"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor"`
}

// Populates Database.DataSources. Pins Notion-Version 2025-09-03.
// See: https://developers.notion.com/reference/retrieve-a-database
func (c *Client) FindDatabaseByIDV2(ctx context.Context, databaseID string) (db Database, err error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/databases/"+databaseID, nil)
	if err != nil {
		return Database{}, fmt.Errorf("notion: invalid request: %w", err)
	}
	req.Header.Set("Notion-Version", "2025-09-03")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Database{}, fmt.Errorf("notion: failed to make HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Database{}, fmt.Errorf("notion: failed to find database: %w", parseErrorResponse(res))
	}

	err = json.NewDecoder(res.Body).Decode(&db)
	if err != nil {
		return Database{}, fmt.Errorf("notion: failed to parse HTTP response: %w", err)
	}

	return db, nil
}

// Pins Notion-Version 2025-09-03.
// See: https://developers.notion.com/reference/query-a-data-source
func (c *Client) QueryDataSource(ctx context.Context, id string, query *DataSourceQuery) (result DataSourceQueryResponse, err error) {
	body := &bytes.Buffer{}

	if query != nil {
		err = json.NewEncoder(body).Encode(query)
		if err != nil {
			return DataSourceQueryResponse{}, fmt.Errorf("notion: failed to encode filter to JSON: %w", err)
		}
	}

	req, err := c.newRequest(ctx, http.MethodPost, fmt.Sprintf("/data_sources/%v/query", id), body)
	if err != nil {
		return DataSourceQueryResponse{}, fmt.Errorf("notion: invalid request: %w", err)
	}
	req.Header.Set("Notion-Version", "2025-09-03")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return DataSourceQueryResponse{}, fmt.Errorf("notion: failed to make HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return DataSourceQueryResponse{}, fmt.Errorf("notion: failed to query data source: %w", parseErrorResponse(res))
	}

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return DataSourceQueryResponse{}, fmt.Errorf("notion: failed to parse HTTP response: %w", err)
	}

	return result, nil
}
