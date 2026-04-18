package github

import (
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	token string
	baseURL string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second}, //reference to the http client with a timeout of 10 seconds
		token: token,
		baseURL: "https://api.github.com",
	}
}

func (c *Client) newRequest(method string, path string) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token) // set the authorization header with the token
	req.Header.Set("Accept", "application/vnd.github+json") 
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28") // set the API version header

	return req, nil
}

