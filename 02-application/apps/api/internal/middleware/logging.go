package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		msg := fmt.Sprintf("%s | %3d | %13v | %15s | %-7s %s",
			requestID, statusCode, latency, clientIP, method, path)

		if statusCode >= 500 {
			gin.DefaultErrorWriter.Write([]byte("[ERROR] " + msg + "\n"))
		} else {
			gin.DefaultWriter.Write([]byte("[INFO] " + msg + "\n"))
		}
	}
}
