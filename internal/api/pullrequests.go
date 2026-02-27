package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// PullRequest represents an Azure DevOps Git pull request.
type PullRequest struct {
	ID           int          `json:"pullRequestId"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Status       string       `json:"status"`
	CreatedBy    IdentityRef  `json:"createdBy"`
	CreationDate string       `json:"creationDate"`
	SourceBranch string       `json:"sourceRefName"`
	TargetBranch string       `json:"targetRefName"`
	MergeStatus  string       `json:"mergeStatus"`
	IsDraft      bool         `json:"isDraft"`
	Repository   PRRepository `json:"repository"`
	Reviewers    []Reviewer   `json:"reviewers"`
	URL          string       `json:"url"`
}

// IdentityRef represents a user identity in Azure DevOps.
type IdentityRef struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
}

// Reviewer represents a pull request reviewer with their vote.
type Reviewer struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	Vote        int    `json:"vote"`
}

// PRRepository is the repository info embedded in a pull request response.
type PRRepository struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Repository represents a Git repository in Azure DevOps.
type Repository struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type pullRequestList struct {
	Count int           `json:"count"`
	Value []PullRequest `json:"value"`
}

type repositoryList struct {
	Count int          `json:"count"`
	Value []Repository `json:"value"`
}

// PullRequestQuery holds search criteria for listing pull requests.
type PullRequestQuery struct {
	Status   string
	Creator  string
	Reviewer string
	Top      int
}

// CreatePRInput holds the fields for creating a new pull request.
type CreatePRInput struct {
	SourceRefName string        `json:"sourceRefName"`
	TargetRefName string        `json:"targetRefName"`
	Title         string        `json:"title"`
	Description   string        `json:"description,omitempty"`
	IsDraft       bool          `json:"isDraft,omitempty"`
	Reviewers     []IdentityRef `json:"reviewers,omitempty"`
}

// ConnectionData represents the response from the connectionData endpoint.
type ConnectionData struct {
	AuthenticatedUser IdentityRef `json:"authenticatedUser"`
}

// ListRepositories returns all Git repositories in a project.
func (c *Client) ListRepositories(project string) ([]Repository, error) {
	rawURL := c.ProjectURL(project, "git/repositories")
	resp, err := c.doRaw(http.MethodGet, rawURL, "application/json", nil)
	if err != nil {
		return nil, err
	}
	var result repositoryList
	if err := decodeOrClose(resp, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// ListPullRequests lists pull requests, optionally scoped to a repository.
// If repoID is empty, lists across all repositories in the project.
func (c *Client) ListPullRequests(project, repoID string, query PullRequestQuery) ([]PullRequest, error) {
	var path string
	if repoID != "" {
		path = fmt.Sprintf("git/repositories/%s/pullrequests", repoID)
	} else {
		path = "git/pullrequests"
	}
	rawURL := c.ProjectURL(project, path)

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}
	q := u.Query()
	if query.Status != "" {
		q.Set("searchCriteria.status", query.Status)
	}
	if query.Creator != "" {
		q.Set("searchCriteria.creatorId", query.Creator)
	}
	if query.Reviewer != "" {
		q.Set("searchCriteria.reviewerId", query.Reviewer)
	}
	if query.Top > 0 {
		q.Set("$top", strconv.Itoa(query.Top))
	}
	u.RawQuery = q.Encode()

	resp, err := c.doRaw(http.MethodGet, u.String(), "application/json", nil)
	if err != nil {
		return nil, err
	}
	var result pullRequestList
	if err := decodeOrClose(resp, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetPullRequest retrieves a single pull request by ID.
func (c *Client) GetPullRequest(project string, id int) (*PullRequest, error) {
	rawURL := c.ProjectURL(project, fmt.Sprintf("git/pullrequests/%d", id))
	resp, err := c.doRaw(http.MethodGet, rawURL, "application/json", nil)
	if err != nil {
		return nil, err
	}
	var pr PullRequest
	if err := decodeOrClose(resp, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// CreatePullRequest creates a new pull request in the given repository.
func (c *Client) CreatePullRequest(project, repoID string, input CreatePRInput) (*PullRequest, error) {
	rawURL := c.ProjectURL(project, fmt.Sprintf("git/repositories/%s/pullrequests", repoID))
	resp, err := c.doRaw(http.MethodPost, rawURL, "application/json", input)
	if err != nil {
		return nil, err
	}
	var pr PullRequest
	if err := decodeOrClose(resp, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// VotePullRequest sets a reviewer's vote on a pull request.
func (c *Client) VotePullRequest(project, repoID string, prID int, reviewerID string, vote int) error {
	path := fmt.Sprintf("git/repositories/%s/pullrequests/%d/reviewers/%s", repoID, prID, reviewerID)
	rawURL := c.ProjectURL(project, path)
	body := map[string]int{"vote": vote}
	resp, err := c.doRaw(http.MethodPut, rawURL, "application/json", body)
	if err != nil {
		return err
	}
	return decodeOrClose(resp, nil)
}

// GetConnectionData returns information about the authenticated user.
func (c *Client) GetConnectionData() (*ConnectionData, error) {
	var data ConnectionData
	if err := c.Get("connectionData", &data); err != nil {
		return nil, err
	}
	return &data, nil
}
