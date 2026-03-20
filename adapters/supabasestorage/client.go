package supabasestorage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL     string
	serviceRole string
	bucket      string
	httpClient  *http.Client
}

func NewClient(baseURL, serviceRoleKey, bucket string, timeout time.Duration) *Client {
	return &Client{
		baseURL:     strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		serviceRole: strings.TrimSpace(serviceRoleKey),
		bucket:      strings.TrimSpace(bucket),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) PutObject(ctx context.Context, path, contentType string, body io.Reader, size int64) error {
	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, url.PathEscape(c.bucket), escapeObjectPath(path))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceRole)
	req.Header.Set("apikey", c.serviceRole)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "false")
	if size >= 0 {
		req.ContentLength = size
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage upload failed: status=%d body=%s", resp.StatusCode, string(data))
	}
	return nil
}

func (c *Client) DeleteObject(ctx context.Context, path string) error {
	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, url.PathEscape(c.bucket), escapeObjectPath(path))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceRole)
	req.Header.Set("apikey", c.serviceRole)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusNotFound {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage delete failed: status=%d body=%s", resp.StatusCode, string(data))
	}
	return nil
}

func escapeObjectPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}
