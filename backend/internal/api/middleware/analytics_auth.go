package middleware

import (
	"net/http"
	"strings"

	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// AnalyticsAuth returns a middleware that validates JWT tokens for analytics endpoints.
func AnalyticsAuth(analyticsService *service.AnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		username, err := analyticsService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store username in context for handlers
		c.Set("analytics_user", username)
		c.Next()
	}
}
