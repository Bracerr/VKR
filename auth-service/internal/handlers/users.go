package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/middleware"
	"github.com/industrial-sed/auth-service/internal/usecases"
)

// UserHandler HTTP для пользователей тенанта.
type UserHandler struct {
	uc *usecases.UserUC
}

// NewUserHandler конструктор.
func NewUserHandler(uc *usecases.UserUC) *UserHandler {
	return &UserHandler{uc: uc}
}

// Create создаёт пользователя.
// @Summary Создать пользователя (ent_admin)
// @Tags users
// @Accept json
// @Produce json
// @Param body body CreateUserRequest true "данные"
// @Success 201 {object} map[string]string
// @Router /api/v1/users [post]
// @Security BearerAuth
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusUnprocessableEntity, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	actor := middleware.Claims(c)
	id, temp, err := h.uc.CreateUser(c.Request.Context(), actor, req.Username, req.Email, req.Role, req.Password)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":                 id,
		"temporary_password": temp,
		"username":           req.Username + "@" + actor.TenantID,
	})
}

// List список пользователей тенанта.
// @Summary Список пользователей (ent_admin)
// @Tags users
// @Produce json
// @Success 200 {array} models.UserCache
// @Router /api/v1/users [get]
// @Security BearerAuth
func (h *UserHandler) List(c *gin.Context) {
	list, err := h.uc.ListUsers(c.Request.Context(), middleware.Claims(c))
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusInternalServerError, err.Error(), http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateRoles меняет роли.
// @Summary Изменить роли пользователя (ent_admin)
// @Tags users
// @Accept json
// @Param id path string true "Keycloak user id"
// @Param body body UpdateRolesRequest true "роли"
// @Success 204
// @Router /api/v1/users/{id}/roles [put]
// @Security BearerAuth
func (h *UserHandler) UpdateRoles(c *gin.Context) {
	var req UpdateRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusUnprocessableEntity, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := h.uc.UpdateUserRoles(c.Request.Context(), middleware.Claims(c), c.Param("id"), req.Roles); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetPassword задаёт пароль пользователю.
// @Summary Задать пароль пользователя (ent_admin)
// @Tags users
// @Accept json
// @Param id path string true "Keycloak user id"
// @Param body body SetPasswordRequest true "пароль"
// @Success 204
// @Router /api/v1/users/{id}/password [put]
// @Security BearerAuth
func (h *UserHandler) SetPassword(c *gin.Context) {
	var req SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusUnprocessableEntity, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := h.uc.SetUserPassword(c.Request.Context(), middleware.Claims(c), c.Param("id"), req.Password); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

// Delete удаляет пользователя.
// @Summary Удалить пользователя (ent_admin)
// @Tags users
// @Param id path string true "Keycloak user id"
// @Success 204
// @Router /api/v1/users/{id} [delete]
// @Security BearerAuth
func (h *UserHandler) Delete(c *gin.Context) {
	if err := h.uc.DeleteUser(c.Request.Context(), middleware.Claims(c), c.Param("id")); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}
