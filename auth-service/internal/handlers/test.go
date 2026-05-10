package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/ports"
	"github.com/industrial-sed/auth-service/internal/usecases"
)

// TestHandler тестовые ручки (pytest).
type TestHandler struct {
	kc      ports.KeycloakClient
	tenants usecases.TenantRepository
	users   usecases.UserCacheRepository
}

// NewTestHandler конструктор.
func NewTestHandler(kc ports.KeycloakClient, tenants usecases.TenantRepository, users usecases.UserCacheRepository) *TestHandler {
	return &TestHandler{kc: kc, tenants: tenants, users: users}
}

// Login password grant для e2e.
// @Summary Тестовый логин (password grant)
// @Tags internal-test
// @Accept json
// @Produce json
// @Param X-Test-Secret header string true "секрет"
// @Param body body TestLoginRequest true "учётные данные"
// @Success 200 {object} map[string]string
// @Router /api/v1/internal/test/login [post]
func (h *TestHandler) Login(c *gin.Context) {
	var req TestLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusUnprocessableEntity, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	tokens, err := h.kc.PasswordGrant(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err.Error(), http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"id_token":      tokens.IDToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// Cleanup удаляет тестовые тенанты и пользователей.
// @Summary Очистка тестовых данных
// @Tags internal-test
// @Param X-Test-Secret header string true "секрет"
// @Param prefix query string false "префикс кода тенанта (по умолчанию test_)"
// @Success 204
// @Router /api/v1/internal/test/cleanup [delete]
func (h *TestHandler) Cleanup(c *gin.Context) {
	prefix := c.Query("prefix")
	if prefix == "" {
		prefix = "test_"
	}
	if err := usecases.TestCleanup(c.Request.Context(), prefix, h.kc, h.tenants, h.users); err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}
