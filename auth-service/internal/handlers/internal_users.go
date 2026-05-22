package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/usecases"
)

// InternalUsersHandler service API для списка пользователей тенанта.
type InternalUsersHandler struct {
	users *usecases.UserUC
}

// NewInternalUsersHandler конструктор.
func NewInternalUsersHandler(users *usecases.UserUC) *InternalUsersHandler {
	return &InternalUsersHandler{users: users}
}

// ListTenantUsers GET /api/v1/internal/tenants/:code/users
func (h *InternalUsersHandler) ListTenantUsers(c *gin.Context) {
	code := c.Param("code")
	list, err := h.users.ListUsersByTenantCode(c.Request.Context(), code)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusInternalServerError, err.Error(), http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, list)
}

// RepairPasswordsRequest тело POST repair-passwords.
type RepairPasswordsRequest struct {
	Password string `json:"password" binding:"required"`
}

// RepairPasswords POST /api/v1/internal/tenants/:code/repair-passwords
// Синхронизирует пароли Keycloak с кэшем (после сбоя seed / required actions).
func (h *InternalUsersHandler) RepairPasswords(c *gin.Context) {
	var req RepairPasswordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusUnprocessableEntity, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	n, err := h.users.RepairTenantPasswords(c.Request.Context(), c.Param("code"), req.Password)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusInternalServerError, err.Error(), http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"repaired": n})
}
