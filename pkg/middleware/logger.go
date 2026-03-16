package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware returns custom logger middleware
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method

		// Log format
		if query != "" {
			path = path + "?" + query
		}

		// Color coding for status
		var statusColor string
		switch {
		case statusCode >= 200 && statusCode < 300:
			statusColor = "\033[32m" // Green
		case statusCode >= 300 && statusCode < 400:
			statusColor = "\033[33m" // Yellow
		case statusCode >= 400 && statusCode < 500:
			statusColor = "\033[31m" // Red
		case statusCode >= 500:
			statusColor = "\033[35m" // Magenta
		}
		resetColor := "\033[0m"

		fmt.Printf("[GIN] %s | %s%3d%s | %13v | %15s | %-7s %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusColor, statusCode, resetColor,
			latency,
			c.ClientIP(),
			method,
			path,
		)
	}
}
