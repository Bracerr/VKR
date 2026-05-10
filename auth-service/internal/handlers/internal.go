package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/middleware"
	"github.com/industrial-sed/auth-service/internal/usecases"
)

// UserInfo внутренний userinfo для микросервисов.
// @Summary Userinfo (X-Service-Secret + JWT пользователя)
// @Tags internal
// @Produce json
// @Param X-Service-Secret header string true "секрет сервиса"
// @Success 200 {object} usecases.UserInfoDTO
// @Router /api/v1/internal/userinfo [get]
func UserInfo(c *gin.Context) {
	dto := usecases.UserInfoFromClaims(middleware.Claims(c))
	if dto == nil {
		RespondError(c, http.StatusUnauthorized, "нет данных", http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, dto)
}
