package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs HTTP requests with structured information
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		requestID := c.GetString("request_id")

		log.Printf("[%s] %s %s - %d (%dms) request_id=%s",
			method, path, c.ClientIP(), status, duration.Milliseconds(), requestID)
	}
}
