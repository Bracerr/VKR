package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/warehouse-service/internal/middleware"
	"github.com/industrial-sed/warehouse-service/internal/models"
)

func tenant(c *gin.Context) (string, bool) {
	cl := middleware.Claims(c)
	if cl == nil || cl.TenantID == "" {
		RespondError(c, http.StatusUnauthorized, "нет tenant_id", http.StatusUnauthorized)
		return "", false
	}
	return cl.TenantID, true
}

func userName(c *gin.Context) string {
	cl := middleware.Claims(c)
	if cl == nil {
		return ""
	}
	if cl.Username != "" {
		return cl.Username
	}
	return cl.Sub
}

func parseUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	s := c.Param(name)
	if s == "" {
		s = c.Query(name)
	}
	id, err := uuid.Parse(s)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "некорректный uuid", http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

func optionalUUIDQuery(c *gin.Context, q string) *uuid.UUID {
	s := c.Query(q)
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func claims(c *gin.Context) *models.Claims {
	return middleware.Claims(c)
}
