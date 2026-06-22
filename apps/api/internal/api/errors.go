package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/soumabali/vexa/internal/middleware"
)

// CORS is an exported wrapper for middleware.CORS to allow access from tests
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return middleware.CORS(allowedOrigins)
}

// APIError represents a structured API error response
type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

var (
	ErrBadRequest          = NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request")
	ErrUnauthorized        = NewAPIError(http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
	ErrForbidden           = NewAPIError(http.StatusForbidden, "FORBIDDEN", "Access denied")
	ErrNotFound            = NewAPIError(http.StatusNotFound, "NOT_FOUND", "Resource not found")
	ErrConflict            = NewAPIError(http.StatusConflict, "CONFLICT", "Resource conflict")
	ErrInternalServerError = NewAPIError(http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred. Please try again later.")
	ErrTooManyRequests     = NewAPIError(http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests")
)

// ErrorHandler is the central error handler middleware.
// It ensures NO internal details leak to clients in production.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			var apiErr *APIError
			if ok := errors.As(err.Err, &apiErr); ok {
				c.JSON(apiErr.Status, gin.H{
					"error":   apiErr.Code,
					"message": apiErr.Message,
				})
			} else {
				// In debug mode, still return generic message to clients but log details
				if gin.Mode() == gin.DebugMode {
					// Log detailed error internally
					gin.DefaultErrorWriter.Write([]byte("[ERROR DETAIL] " + err.Error() + "\n"))
				}

				// Always return generic message to client - never expose stack traces or internal details
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "INTERNAL_ERROR",
					"message": "An internal error occurred. Please try again later.",
				})
			}
		}
	}
}

func ValidationError(err error) *APIError {
	return NewAPIError(http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
}
