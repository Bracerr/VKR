package usecases

import (
	"context"
	"sync"
	"time"

	"github.com/industrial-sed/auth-service/internal/ports"
)

// adminTokenSource кэширует токен Keycloak Admin API (master).
type adminTokenSource struct {
	kc    ports.KeycloakClient
	mu    sync.Mutex
	token string
	exp   time.Time
}

func newAdminTokenSource(kc ports.KeycloakClient) *adminTokenSource {
	return &adminTokenSource{kc: kc}
}

func (a *adminTokenSource) Token(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && time.Now().Before(a.exp) {
		return a.token, nil
	}
	jwt, err := a.kc.LoginAdmin(ctx)
	if err != nil {
		return "", err
	}
	ttl := 50 * time.Second
	if jwt.ExpiresIn > 0 && jwt.ExpiresIn > 30 {
		ttl = time.Duration(jwt.ExpiresIn-20) * time.Second
	}
	a.token = jwt.AccessToken
	a.exp = time.Now().Add(ttl)
	return a.token, nil
}

// Проверка типа для моков.
var _ interface {
	Token(ctx context.Context) (string, error)
} = (*adminTokenSource)(nil)
