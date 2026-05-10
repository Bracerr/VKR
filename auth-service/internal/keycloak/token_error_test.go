package keycloak_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/industrial-sed/auth-service/internal/keycloak"
)

func TestPublicTokenErrorMessage(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "account not ready",
			status: 400,
			body:   `{"error":"invalid_grant","error_description":"Account is not fully set up"}`,
			want:   "Вход невозможен: учётная запись не готова к входу. Завершите обязательные действия (например, смену временного пароля или заполнение профиля) или обратитесь к администратору.",
		},
		{
			name:   "bad credentials",
			status: 401,
			body:   `{"error":"invalid_grant","error_description":"Invalid user credentials"}`,
			want:   "Неверный логин или пароль.",
		},
		{
			name:   "server error",
			status: 502,
			body:   `bad gateway`,
			want:   "Сервис авторизации временно недоступен. Попробуйте позже.",
		},
		{
			name:   "empty body",
			status: 400,
			body:   "",
			want:   "Не удалось выполнить вход. Попробуйте позже.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := keycloak.PublicTokenErrorMessage(tc.status, []byte(tc.body))
			require.Equal(t, tc.want, got)
		})
	}
}
