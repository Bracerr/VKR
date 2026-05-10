package usecases

import (
	"context"

	"github.com/industrial-sed/auth-service/internal/models"
)

// TenantRepository — хранение кэша предприятий.
type TenantRepository interface {
	Create(ctx context.Context, t *models.Tenant) error
	List(ctx context.Context) ([]models.Tenant, error)
	GetByCode(ctx context.Context, code string) (*models.Tenant, error)
	Delete(ctx context.Context, code string) error
	ListByCodePrefix(ctx context.Context, prefix string) ([]models.Tenant, error)
}

// UserCacheRepository — кэш пользователей.
type UserCacheRepository interface {
	Upsert(ctx context.Context, u *models.UserCache) error
	ListByTenant(ctx context.Context, tenantCode string) ([]models.UserCache, error)
	GetByKeycloakID(ctx context.Context, keycloakID string) (*models.UserCache, error)
	Delete(ctx context.Context, keycloakID string) error
	DeleteByTenant(ctx context.Context, tenantCode string) error
}

// Notifier — уведомления (mock / Kafka).
type Notifier interface {
	NotifyUserCreated(ctx context.Context, payload UserCreatedPayload) error
}

// UserCreatedPayload — событие создания пользователя.
type UserCreatedPayload struct {
	TenantCode        string
	Username          string
	Email             string
	TemporaryPassword string
	KeycloakUserID    string
}
