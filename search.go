package notion

import (
	"encoding/json"
	"fmt"
)

type SearchOpts struct {
	Query       string        `json:"query,omitempty"`
	Sort        *SearchSort   `json:"sort,omitempty"`
	Filter      *SearchFilter `json:"filter,omitempty"`
	StartCursor string        `json:"start_cursor,omitempty"`
	PageSize    int           `json:"page_size,omitempty"`
}

type SearchSort struct {
	Direction SortDirection       `json:"direction,omitempty"`
	Timestamp SearchSortTimestamp `json:"timestamp"`
}

type SearchSortTimestamp string

type SearchFilter struct {
	Value    string `json:"value"`
	Property string `json:"property"`
}

type SearchResponse struct {
	// Results are either pages or data sources. See `SearchResults.UnmarshalJSON`.
	Results    SearchResults `json:"results"`
	HasMore    bool          `json:"has_more"`
	NextCursor *string       `json:"next_cursor"`
}

type SearchResults []interface{}

const SearchSortTimestampLastEditedTime SearchSortTimestamp = "last_edited_time"

func (sr *SearchResults) UnmarshalJSON(b []byte) error {
	rawResults := []json.RawMessage{}
	err := json.Unmarshal(b, &rawResults)
	if err != nil {
		return err
	}

	type Object struct {
		Object string `json:"object"`
	}

	results := make(SearchResults, len(rawResults))

	for i, rawResult := range rawResults {
		obj := Object{}
		err := json.Unmarshal(rawResult, &obj)
		if err != nil {
			return err
		}

		switch obj.Object {
		case "data_source":
			var ds DataSource
			err := json.Unmarshal(rawResult, &ds)
			if err != nil {
				return err
			}
			results[i] = ds
		case "page":
			var page Page
			err := json.Unmarshal(rawResult, &page)
			if err != nil {
				return err
			}
			results[i] = page
		default:
			return fmt.Errorf("unsupported result object %q", obj.Object)
		}
	}

	*sr = results

	return nil
}
