package httpx

import "github.com/gin-gonic/gin"

// ErrorBody единый формат ошибки API.
type ErrorBody struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// ErrorJSON пишет JSON-ошибку.
func ErrorJSON(c *gin.Context, httpStatus int, message string, code int) {
	c.JSON(httpStatus, ErrorBody{Error: message, Code: code})
}
