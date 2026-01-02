package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := strings.TrimSpace(c.GetHeader("Authorization"))
		if authz == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}

		parts := strings.SplitN(authz, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
			return
		}

		userID, username, err := s.ParseToken(parts[1])
		if err != nil || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// IMPORTANT: handlers expect snake_case keys
		c.Set("user_id", userID)
		c.Set("username", username)

		// Back-compat (some code may use camelCase)
		c.Set("userId", userID)

		c.Next()
	}
}
