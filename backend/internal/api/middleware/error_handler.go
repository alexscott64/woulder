package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler centralizes error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			requestID := c.GetString("request_id")

			// Determine status code
			status := http.StatusInternalServerError
			message := "Internal server error"

			if errors.Is(err, sql.ErrNoRows) {
				status = http.StatusNotFound
				message = "Resource not found"
			} else if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusRequestTimeout
				message = "Request timeout"
			}

			c.JSON(status, gin.H{
				"error":      message,
				"request_id": requestID,
			})
		}
	}
}
