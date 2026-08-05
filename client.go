package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yinebebt/deeplink"
)

// Client is an HTTP client for a deeplink server.
type Client struct {
	base   string
	apiKey string
	http   *http.Client
}

// createResponse is the POST /shorten response body.
type createResponse struct {
	ShortURL string `json:"short_url"`
}

func newClient(endpoint, apiKey string) (*Client, error) {
	base := strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if base == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if _, err := url.ParseRequestURI(base); err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}
	return &Client{
		base:   base,
		apiKey: apiKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) CreateLink(ctx context.Context, in deeplink.Link) (*deeplink.LinkResponse, error) {
	var created createResponse
	if err := c.do(ctx, http.MethodPost, "/shorten", in, &created); err != nil {
		return nil, err
	}
	shortID, err := shortIDFromURL(created.ShortURL)
	if err != nil {
		return nil, err
	}
	return c.GetLink(ctx, in.Type, shortID)
}

func (c *Client) GetLink(ctx context.Context, linkType, shortID string) (*deeplink.LinkResponse, error) {
	path := fmt.Sprintf("/links/%s/%s", url.PathEscape(linkType), url.PathEscape(shortID))
	var out deeplink.LinkResponse
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateLink(ctx context.Context, shortID string, patch map[string]any) (*deeplink.LinkResponse, error) {
	path := "/" + url.PathEscape(shortID)
	var out deeplink.LinkResponse
	if err := c.do(ctx, http.MethodPatch, path, patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteLink(ctx context.Context, shortID string) error {
	return c.do(ctx, http.MethodDelete, "/"+url.PathEscape(shortID), nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = res.Status
		}
		return &apiError{Status: res.StatusCode, Message: msg}
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

type apiError struct {
	Status  int
	Message string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("deeplink API %d: %s", e.Status, e.Message)
}

func (e *apiError) NotFound() bool {
	return e.Status == http.StatusNotFound
}

func shortIDFromURL(shortURL string) (string, error) {
	u, err := url.Parse(shortURL)
	if err != nil {
		return "", fmt.Errorf("parse short_url: %w", err)
	}
	id := strings.Trim(u.Path, "/")
	if id == "" {
		return "", fmt.Errorf("short_url missing path: %q", shortURL)
	}
	if i := strings.LastIndex(id, "/"); i >= 0 {
		id = id[i+1:]
	}
	return id, nil
}
