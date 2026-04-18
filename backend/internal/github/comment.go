package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateIssueCommentRequest represents the JSON body used to create
// a PR comment through GitHub's issues comments API
type CreateIssueCommentRequest struct {
	Body string `json:"body"`
}

// IssueCommentResponse represents the subset of GitHub's comment response
// that we care about for now
type IssueCommentResponse struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
	URL  string `json:"html_url"`
}

// PostPullRequestComment creates a comment on a pull request
func (c *Client) PostPullRequestComment(owner string, repo string, prNumber int, body string) (*IssueCommentResponse, error) {
	// Build the API path for issue comments. Note that PRs use the same numbering space as issues for this endpoint
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)

	// prepare the JSON request body
	payload := CreateIssueCommentRequest{
		Body: body,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Build the full GitHub API URL
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	// add the required GitHub headers
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// GitHub should return 201 Created when the comment is created.
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var comment IssueCommentResponse
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, err
	}

	return &comment, nil
}