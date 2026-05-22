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
