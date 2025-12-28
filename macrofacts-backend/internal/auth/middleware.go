package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/httpapi"
)

func (s *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			httpapi.WriteError(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token", nil)
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))

		uid, un, err := s.ParseToken(raw)
		if err != nil {
			httpapi.WriteError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", nil)
			return
		}

		c.Set("userId", uid)
		c.Set("username", un)
		c.Next()
	}
}
