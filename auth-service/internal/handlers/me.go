package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/middleware"
	"github.com/industrial-sed/auth-service/internal/usecases"
)

// Me текущий пользователь из JWT.
// @Summary Профиль по JWT
// @Tags auth
// @Produce json
// @Success 200 {object} usecases.UserInfoDTO
// @Router /api/v1/auth/me [get]
// @Security BearerAuth
func Me(c *gin.Context) {
	dto := usecases.UserInfoFromClaims(middleware.Claims(c))
	if dto == nil {
		RespondError(c, http.StatusUnauthorized, "нет данных", http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, dto)
}
