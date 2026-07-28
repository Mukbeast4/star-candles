package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const maxStargazerPages = 400

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		token:      token,
	}
}

func (c *Client) StarCount(ctx context.Context, repo string) (int, error) {
	body, err := c.get(ctx, fmt.Sprintf("/repos/%s", repo), "application/vnd.github+json")
	if err != nil {
		return 0, fmt.Errorf("fetch repo: %w", err)
	}

	var payload struct {
		StargazersCount int `json:"stargazers_count"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, fmt.Errorf("decode repo: %w", err)
	}
	return payload.StargazersCount, nil
}

func (c *Client) StarEvents(ctx context.Context, repo string) ([]time.Time, error) {
	var events []time.Time
	for page := 1; page <= maxStargazerPages; page++ {
		path := fmt.Sprintf("/repos/%s/stargazers?per_page=100&page=%d", repo, page)
		body, err := c.get(ctx, path, "application/vnd.github.star+json")
		if err != nil {
			return nil, fmt.Errorf("fetch stargazers page %d: %w", page, err)
		}

		var payload []struct {
			StarredAt time.Time `json:"starred_at"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode stargazers page %d: %w", page, err)
		}

		for _, s := range payload {
			events = append(events, s.StarredAt)
		}
		if len(payload) < 100 {
			break
		}
	}

	sort.Slice(events, func(i, j int) bool { return events[i].Before(events[j]) })
	return events, nil
}

func (c *Client) get(ctx context.Context, path, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("github api %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}
