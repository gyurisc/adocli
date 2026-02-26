// Package api provides a REST client for the Azure DevOps API.
package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultAPIVersion = "7.1"

// Client is an Azure DevOps REST API client.
type Client struct {
	BaseURL    string
	PAT        string
	APIVersion string
	HTTP       *http.Client
}

// NewClient creates a Client for the given Azure DevOps organization.
// The base URL is https://dev.azure.com/{org}/_apis.
func NewClient(org, pat string) *Client {
	return &Client{
		BaseURL:    fmt.Sprintf("https://dev.azure.com/%s/_apis", org),
		PAT:        pat,
		APIVersion: defaultAPIVersion,
		HTTP:       &http.Client{},
	}
}

// authHeader returns the Basic auth header value for PAT authentication.
func (c *Client) authHeader() string {
	token := base64.StdEncoding.EncodeToString([]byte(":" + c.PAT))
	return "Basic " + token
}

// do executes an HTTP request and returns the response.
func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := fmt.Sprintf("%s/%s", c.BaseURL, path)

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	// Append api-version query parameter.
	q := req.URL.Query()
	q.Set("api-version", c.APIVersion)
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	return resp, nil
}

// decodeOrClose reads the response body into result. It always closes the body.
func decodeOrClose(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &Error{StatusCode: resp.StatusCode, Body: string(body)}
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Get performs a GET request and decodes the JSON response into result.
func (c *Client) Get(path string, result interface{}) error {
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return decodeOrClose(resp, result)
}

// Post performs a POST request with a JSON body and decodes the response.
func (c *Client) Post(path string, body, result interface{}) error {
	resp, err := c.do(http.MethodPost, path, body)
	if err != nil {
		return err
	}
	return decodeOrClose(resp, result)
}

// Patch performs a PATCH request with a JSON body and decodes the response.
func (c *Client) Patch(path string, body, result interface{}) error {
	resp, err := c.do(http.MethodPatch, path, body)
	if err != nil {
		return err
	}
	return decodeOrClose(resp, result)
}

// ProjectURL constructs a project-scoped API URL.
func (c *Client) ProjectURL(project, path string) string {
	orgBase := strings.TrimSuffix(c.BaseURL, "/_apis")
	return fmt.Sprintf("%s/%s/_apis/%s", orgBase, project, path)
}

// doRaw executes an HTTP request with a caller-specified full URL and content type.
func (c *Client) doRaw(method, rawURL, contentType string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, rawURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", contentType)

	q := req.URL.Query()
	q.Set("api-version", c.APIVersion)
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	return resp, nil
}

// Error represents an error response from the Azure DevOps API.
type Error struct {
	StatusCode int
	Body       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Azure DevOps API error (HTTP %d): %s", e.StatusCode, e.Body)
}
