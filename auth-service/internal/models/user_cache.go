package models

import "time"

// UserCache — дублирование данных пользователя для быстрых запросов.
type UserCache struct {
	KeycloakID string    `json:"keycloak_id"`
	TenantCode string    `json:"tenant_code"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Roles      []string  `json:"roles"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
