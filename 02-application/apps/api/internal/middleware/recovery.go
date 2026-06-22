package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RecoveryHandler() gin.RecoveryFunc {
	return func(c *gin.Context, err interface{}) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		})
		if gin.Mode() == gin.DebugMode {
			fmt.Fprintf(gin.DefaultErrorWriter, "[PANIC] %v\n", err)
		}
	}
}
