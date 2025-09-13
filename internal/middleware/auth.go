package middleware

import (
	"net/http"

	"github.com/iokiris/efm-subscription-api/internal/logger"

	"github.com/gin-gonic/gin"
)

// JWTMiddleware
// NOTE: заглушка.
// Добавлено как возможность быстро накинуть авторизацию при необходимости.
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.L.Info("JWTMiddleware")
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}
		tokenStr := authHeader[len("Bearer "):]

		// нет реализации
		userID := tokenStr

		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return

		}
		c.Set("user_id", userID)
		c.Next()
	}
}
