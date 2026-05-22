package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AuthTenantUser пользователь тенанта из auth-service.
type AuthTenantUser struct {
	KeycloakID string   `json:"keycloak_id"`
	TenantCode string   `json:"tenant_code"`
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
}

// AuthUsersClient internal API auth-service.
type AuthUsersClient struct {
	BaseURL string
	Secret  string
	HTTP    *http.Client
}

// NewAuthUsersClient конструктор.
func NewAuthUsersClient(baseURL, secret string) *AuthUsersClient {
	return &AuthUsersClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Secret:  secret,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// ListTenantUsers GET /api/v1/internal/tenants/:code/users
func (c *AuthUsersClient) ListTenantUsers(ctx context.Context, tenantCode string) ([]AuthTenantUser, error) {
	if c == nil || c.BaseURL == "" || c.Secret == "" {
		return nil, fmt.Errorf("auth users client not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/v1/internal/tenants/"+tenantCode+"/users", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Service-Secret", c.Secret)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth list users: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var list []AuthTenantUser
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, err
	}
	return list, nil
}
