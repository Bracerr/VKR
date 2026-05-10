package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/jwtverify"
	"github.com/industrial-sed/auth-service/internal/middleware"
	"github.com/industrial-sed/auth-service/internal/usecases"
)

// AuthHandler OIDC BFF.
type AuthHandler struct {
	auth   *usecases.AuthUC
	parser *jwtverify.Parser
	secure bool
}

// NewAuthHandler конструктор.
func NewAuthHandler(auth *usecases.AuthUC, parser *jwtverify.Parser, secure bool) *AuthHandler {
	return &AuthHandler{auth: auth, parser: parser, secure: secure}
}

func setTokenCookie(c *gin.Context, name, value string, maxAge int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearTokenCookie(c *gin.Context, name string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// Login редирект на Keycloak (PKCE).
// @Summary Начать вход (OIDC code + PKCE)
// @Tags auth
// @Param return_to query string false "путь на фронте после входа"
// @Success 302
// @Router /api/v1/auth/login [get]
func (h *AuthHandler) Login(c *gin.Context) {
	returnTo := c.Query("return_to")
	url, err := h.auth.BuildAuthorizeURL(returnTo)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Redirect(http.StatusFound, url)
}

// Callback обмен code на токены, cookies, редирект на SPA.
// @Summary OIDC callback
// @Tags auth
// @Success 302
// @Router /api/v1/auth/callback [get]
func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if state == "" {
		RespondError(c, http.StatusBadRequest, "отсутствует параметр state", http.StatusBadRequest)
		return
	}
	tokens, returnTo, err := h.auth.ExchangeCallback(c.Request.Context(), code, state)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err.Error(), http.StatusUnauthorized)
		return
	}
	claims, err := h.parser.ParseAccessToken(c.Request.Context(), tokens.AccessToken)
	if err == nil {
		_ = h.auth.SyncUserCache(c.Request.Context(), claims)
	}
	maxAge := tokens.ExpiresIn
	if maxAge <= 0 {
		maxAge = 300
	}
	refreshMax := tokens.RefreshExpiresIn
	if refreshMax <= 0 {
		refreshMax = 86400 * 7
	}
	setTokenCookie(c, middleware.CookieAccessToken, tokens.AccessToken, maxAge, h.secure)
	setTokenCookie(c, middleware.CookieRefreshToken, tokens.RefreshToken, refreshMax, h.secure)
	if tokens.IDToken != "" {
		setTokenCookie(c, middleware.CookieIDToken, tokens.IDToken, refreshMax, h.secure)
	}
	loc := h.auth.FrontendBase()
	if strings.HasPrefix(returnTo, "/") {
		loc += returnTo
	} else {
		loc += "/" + returnTo
	}
	c.Redirect(http.StatusFound, loc)
}

// Refresh обновляет access по refresh cookie.
// @Summary Обновить access token
// @Tags auth
// @Success 204
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	rt, err := c.Cookie(middleware.CookieRefreshToken)
	if err != nil || rt == "" {
		RespondError(c, http.StatusUnauthorized, "нет refresh", http.StatusUnauthorized)
		return
	}
	tokens, err := h.auth.Refresh(c.Request.Context(), rt)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err.Error(), http.StatusUnauthorized)
		return
	}
	maxAge := tokens.ExpiresIn
	if maxAge <= 0 {
		maxAge = 300
	}
	refreshMax := tokens.RefreshExpiresIn
	if refreshMax <= 0 {
		refreshMax = 86400 * 7
	}
	setTokenCookie(c, middleware.CookieAccessToken, tokens.AccessToken, maxAge, h.secure)
	setTokenCookie(c, middleware.CookieRefreshToken, tokens.RefreshToken, refreshMax, h.secure)
	if tokens.IDToken != "" {
		setTokenCookie(c, middleware.CookieIDToken, tokens.IDToken, refreshMax, h.secure)
	}
	c.Status(http.StatusNoContent)
}

// Logout очищает cookies и возвращает URL Keycloak end_session (клиент может сделать редирект).
// @Summary Выход
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	rt, _ := c.Cookie(middleware.CookieRefreshToken)
	idt, _ := c.Cookie(middleware.CookieIDToken)
	url, _ := h.auth.LogoutSession(c.Request.Context(), rt, idt)
	clearTokenCookie(c, middleware.CookieAccessToken, h.secure)
	clearTokenCookie(c, middleware.CookieRefreshToken, h.secure)
	clearTokenCookie(c, middleware.CookieIDToken, h.secure)
	c.JSON(http.StatusOK, gin.H{"end_session_url": url})
}
