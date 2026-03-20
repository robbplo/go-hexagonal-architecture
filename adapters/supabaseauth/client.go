package supabaseauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type Client struct {
	baseURL      string
	serviceRole  string
	authUserPath string
	adminPath    string
	httpClient   *http.Client
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	User  struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

func NewClient(baseURL, serviceRoleKey, authUserPath, adminPath string, timeout time.Duration) *Client {
	return &Client{
		baseURL:      strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		serviceRole:  strings.TrimSpace(serviceRoleKey),
		authUserPath: ensurePath(authUserPath),
		adminPath:    ensurePath(adminPath),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Authenticate(ctx context.Context, bearerToken string) (ports.AuthenticatedUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+c.authUserPath, nil)
	if err != nil {
		return ports.AuthenticatedUser{}, err
	}
	req.Header.Set("Authorization", bearerToken)
	req.Header.Set("apikey", c.serviceRole)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ports.AuthenticatedUser{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ports.AuthenticatedUser{}, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ports.AuthenticatedUser{}, domainerrors.ErrUnauthorized
	}
	if resp.StatusCode >= 400 {
		return ports.AuthenticatedUser{}, fmt.Errorf("authenticate token: status=%d body=%s", resp.StatusCode, string(body))
	}

	var decoded userResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return ports.AuthenticatedUser{}, err
	}
	id := decoded.ID
	email := decoded.Email
	if id == "" {
		id = decoded.User.ID
		email = decoded.User.Email
	}
	return ports.AuthenticatedUser{
		ID:    id,
		Email: strings.ToLower(strings.TrimSpace(email)),
	}, nil
}

func (c *Client) InviteUser(ctx context.Context, email, redirectURL string) (ports.IdentityUser, error) {
	payload, err := json.Marshal(map[string]any{
		"email":       strings.ToLower(strings.TrimSpace(email)),
		"redirect_to": strings.TrimSpace(redirectURL),
	})
	if err != nil {
		return ports.IdentityUser{}, err
	}
	response, err := c.doAdminRequest(ctx, http.MethodPost, c.adminPath+"/invite", payload)
	if err != nil {
		return ports.IdentityUser{}, err
	}
	return parseIdentityUser(response)
}

func (c *Client) DisableUser(ctx context.Context, userID string) error {
	payload, err := json.Marshal(map[string]any{
		"ban_duration": "876000h",
	})
	if err != nil {
		return err
	}
	_, err = c.doAdminRequest(ctx, http.MethodPut, c.adminPath+"/users/"+strings.TrimSpace(userID), payload)
	return err
}

func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	_, err := c.doAdminRequest(ctx, http.MethodDelete, c.adminPath+"/users/"+strings.TrimSpace(userID), nil)
	return err
}

func (c *Client) CreateUser(ctx context.Context, email, password string, emailConfirmed bool) (ports.IdentityUser, error) {
	payload, err := json.Marshal(map[string]any{
		"email":         strings.ToLower(strings.TrimSpace(email)),
		"password":      password,
		"email_confirm": emailConfirmed,
	})
	if err != nil {
		return ports.IdentityUser{}, err
	}
	response, err := c.doAdminRequest(ctx, http.MethodPost, c.adminPath+"/users", payload)
	if err != nil {
		return ports.IdentityUser{}, err
	}
	return parseIdentityUser(response)
}

func (c *Client) doAdminRequest(ctx context.Context, method, path string, payload []byte) ([]byte, error) {
	var body io.Reader
	if len(payload) > 0 {
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", c.serviceRole)
	req.Header.Set("Authorization", "Bearer "+c.serviceRole)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return data, nil
	case http.StatusNotFound:
		return nil, domainerrors.ErrNotFound
	case http.StatusConflict:
		return nil, domainerrors.ErrAlreadyExists
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, domainerrors.ErrUnauthorized
	default:
		return nil, fmt.Errorf("supabase admin request failed: status=%d body=%s", resp.StatusCode, string(data))
	}
}

func parseIdentityUser(data []byte) (ports.IdentityUser, error) {
	if len(data) == 0 {
		return ports.IdentityUser{}, nil
	}
	var decoded userResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		return ports.IdentityUser{}, err
	}
	id := decoded.ID
	email := decoded.Email
	if id == "" {
		id = decoded.User.ID
		email = decoded.User.Email
	}
	return ports.IdentityUser{
		ID:    id,
		Email: strings.ToLower(strings.TrimSpace(email)),
	}, nil
}

func ensurePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}
