package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/keycloak"
	"github.com/industrial-sed/auth-service/internal/middleware"
	"github.com/industrial-sed/auth-service/internal/usecases"
)

// TenantHandler HTTP для тенантов.
type TenantHandler struct {
	uc   *usecases.TenantUC
	user *usecases.UserUC // для bootstrap ent_admin
}

// NewTenantHandler конструктор.
func NewTenantHandler(uc *usecases.TenantUC, user *usecases.UserUC) *TenantHandler {
	return &TenantHandler{uc: uc, user: user}
}

// BootstrapEntAdminRequest первый админ предприятия.
type BootstrapEntAdminRequest struct {
	Username string `json:"username" binding:"required,min=1,max=64"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// Create создаёт предприятие.
// @Summary Создать предприятие (super_admin)
// @Tags tenants
// @Accept json
// @Produce json
// @Param body body CreateTenantRequest true "данные"
// @Success 201 {object} models.Tenant
// @Failure 400 {object} httpx.ErrorBody
// @Router /api/v1/tenants [post]
// @Security BearerAuth
func (h *TenantHandler) Create(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusUnprocessableEntity, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	t, err := h.uc.CreateTenant(c.Request.Context(), req.Code, req.Name)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, t)
}

// List список предприятий.
// @Summary Список предприятий (super_admin)
// @Tags tenants
// @Produce json
// @Success 200 {array} models.Tenant
// @Router /api/v1/tenants [get]
// @Security BearerAuth
func (h *TenantHandler) List(c *gin.Context) {
	list, err := h.uc.ListTenants(c.Request.Context())
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error(), http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, list)
}

// Delete удаляет предприятие.
// @Summary Удалить предприятие (super_admin)
// @Tags tenants
// @Param code path string true "код предприятия"
// @Success 204
// @Router /api/v1/tenants/{code} [delete]
// @Security BearerAuth
func (h *TenantHandler) Delete(c *gin.Context) {
	code := c.Param("code")
	if err := h.uc.DeleteTenant(c.Request.Context(), code); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

// BootstrapEntAdmin создаёт первого ent_admin в предприятии (super_admin).
// @Summary Bootstrap первого администратора предприятия
// @Tags tenants
// @Accept json
// @Param code path string true "код тенанта"
// @Param body body BootstrapEntAdminRequest true "данные"
// @Success 201 {object} map[string]string
// @Router /api/v1/tenants/{code}/ent-admin [post]
// @Security BearerAuth
func (h *TenantHandler) BootstrapEntAdmin(c *gin.Context) {
	if h.user == nil {
		RespondError(c, http.StatusInternalServerError, "user usecase not configured", http.StatusInternalServerError)
		return
	}
	cl := middleware.Claims(c)
	if cl == nil || !cl.HasRole(keycloak.RoleSuperAdmin) {
		RespondError(c, http.StatusForbidden, "только super_admin", http.StatusForbidden)
		return
	}
	var req BootstrapEntAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusUnprocessableEntity, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	id, err := h.user.CreateEntAdminBySuper(c.Request.Context(), c.Param("code"), req.Username, req.Email, req.Password)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "username": req.Username + "@" + c.Param("code")})
}
