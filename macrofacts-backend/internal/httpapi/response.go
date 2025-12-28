package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	Details   map[string]any `json:"details,omitempty"`
}

func RequestID(c *gin.Context) string {
	if v, ok := c.Get("request_id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func WriteError(c *gin.Context, status int, code, message string, details map[string]any) {
	c.AbortWithStatusJSON(status, ErrorEnvelope{Error: APIError{
		Code:      code,
		Message:   message,
		RequestID: RequestID(c),
		Details:   details,
	}})
}

func BadRequest(c *gin.Context, message string, details map[string]any) {
	WriteError(c, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

func Unauthorized(c *gin.Context, message string) {
	WriteError(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func NotFound(c *gin.Context, message string, details map[string]any) {
	WriteError(c, http.StatusNotFound, "NOT_FOUND", message, details)
}

func Internal(c *gin.Context, message string) {
	WriteError(c, http.StatusInternalServerError, "INTERNAL", message, nil)
}
