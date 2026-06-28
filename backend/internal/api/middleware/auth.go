package middleware

import (
	"net/http"
	"strings"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserID    = "user_id"
	ContextUserEmail = "user_email"
	ContextUserRole  = "user_role"
)

func Auth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		claims, err := authService.ValidateAccessToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserEmail, claims.Email)
		c.Set(ContextUserRole, claims.Role)
		c.Next()
	}
}

func RequireMoneyWrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ContextUserRole)
		if !models.CanWriteMoney(roleString(role)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) models.CurrentUser {
	return models.CurrentUser{ID: roleString(mustGet(c, ContextUserID)), Email: roleString(mustGet(c, ContextUserEmail)), Role: roleString(mustGet(c, ContextUserRole))}
}

func mustGet(c *gin.Context, key string) interface{} { v, _ := c.Get(key); return v }
func roleString(v interface{}) string                { s, _ := v.(string); return s }
