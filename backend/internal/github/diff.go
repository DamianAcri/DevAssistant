package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PullRequestFile represents the subset of GitHub PR file data we need.
type PullRequestFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
	Patch     string `json:"patch"`
}

// GetPullRequestDiff fetches the changed files of a PR and turns them
// into a single diff-like string that can later be sent to an LLM
func (c *Client) GetPullRequestDiff(owner string, repo string, prNumber int) (string, error) {
	// Build the API path for PR files
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/files", owner, repo, prNumber)

	req, err := c.newRequest(http.MethodGet, path)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// GitHub should return 200 here
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github API returned %d: %s", resp.StatusCode, string(body))
	}

	var files []PullRequestFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return "", err
	}

	// Build a readable string
	var builder strings.Builder

	for _, file := range files {
		// Add file metadata first.
		builder.WriteString(fmt.Sprintf("File: %s\n", file.Filename))
		builder.WriteString(fmt.Sprintf("Status: %s\n", file.Status))
		builder.WriteString(fmt.Sprintf("Additions: %d\n", file.Additions))
		builder.WriteString(fmt.Sprintf("Deletions: %d\n", file.Deletions))
		builder.WriteString(fmt.Sprintf("Changes: %d\n", file.Changes))

		// add the patch only if GitHub provided one (it may not for large files or binary files)
		if file.Patch != "" {
			builder.WriteString("Patch:\n")
			builder.WriteString(file.Patch)
			builder.WriteString("\n")
		} else {
			builder.WriteString("Patch: <not available>\n")
		}

		builder.WriteString("\n---\n\n")
	}

	return builder.String(), nil
}