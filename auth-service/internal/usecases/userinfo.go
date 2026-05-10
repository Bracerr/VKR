package usecases

import (
	"github.com/industrial-sed/auth-service/internal/models"
)

// UserInfoDTO ответ для других микросервисов.
type UserInfoDTO struct {
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id"`
	Username string   `json:"username"`
	Email    string   `json:"email,omitempty"`
	Roles    []string `json:"roles"`
}

// UserInfoFromClaims строит DTO из JWT-claims.
func UserInfoFromClaims(c *models.Claims) *UserInfoDTO {
	if c == nil {
		return nil
	}
	return &UserInfoDTO{
		UserID:   c.Sub,
		TenantID: c.TenantID,
		Username: c.Username,
		Email:    c.Email,
		Roles:    c.RealmRoles,
	}
}
