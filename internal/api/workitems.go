package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// WiqlResult is the response from a WIQL query.
type WiqlResult struct {
	WorkItems []WiqlWorkItemRef `json:"workItems"`
}

// WiqlWorkItemRef is a lightweight reference returned by WIQL.
type WiqlWorkItemRef struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

// WorkItem represents a work item from the API.
type WorkItem struct {
	ID     int                    `json:"id"`
	Rev    int                    `json:"rev"`
	Fields map[string]interface{} `json:"fields"`
	URL    string                 `json:"url"`
}

// WorkItemList is the response when fetching multiple work items.
type WorkItemList struct {
	Count int        `json:"count"`
	Value []WorkItem `json:"value"`
}

// PatchField represents a single JSON Patch operation for work item create/update.
type PatchField struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// QueryByWiql runs a WIQL query and returns matching work item references.
// The top parameter limits the number of results returned by the server.
func (c *Client) QueryByWiql(project, wiql string, top int) (*WiqlResult, error) {
	body := map[string]string{"query": wiql}
	url := c.ProjectURL(project, fmt.Sprintf("wit/wiql?$top=%d", top))

	resp, err := c.doRaw(http.MethodPost, url, "application/json", body)
	if err != nil {
		return nil, err
	}
	var result WiqlResult
	if err := decodeOrClose(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWorkItem retrieves a single work item by ID.
func (c *Client) GetWorkItem(project string, id int) (*WorkItem, error) {
	path := fmt.Sprintf("wit/workitems/%d", id)
	var wi WorkItem
	if err := c.Get(path, &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}

// GetWorkItems retrieves multiple work items by IDs in a single batch call.
func (c *Client) GetWorkItems(project string, ids []int) ([]WorkItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.Itoa(id)
	}
	path := fmt.Sprintf("wit/workitems?ids=%s", strings.Join(strs, ","))
	var result WorkItemList
	if err := c.Get(path, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// CreateWorkItem creates a new work item using JSON Patch.
func (c *Client) CreateWorkItem(project, workItemType string, fields []PatchField) (*WorkItem, error) {
	url := c.ProjectURL(project, fmt.Sprintf("wit/workitems/$%s", workItemType))

	resp, err := c.doRaw(http.MethodPost, url, "application/json-patch+json", fields)
	if err != nil {
		return nil, err
	}
	var wi WorkItem
	if err := decodeOrClose(resp, &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}

// UpdateWorkItem updates an existing work item using JSON Patch.
func (c *Client) UpdateWorkItem(project string, id int, fields []PatchField) (*WorkItem, error) {
	rawURL := fmt.Sprintf("%s/wit/workitems/%d", c.BaseURL, id)

	resp, err := c.doRaw(http.MethodPatch, rawURL, "application/json-patch+json", fields)
	if err != nil {
		return nil, err
	}
	var wi WorkItem
	if err := decodeOrClose(resp, &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}
