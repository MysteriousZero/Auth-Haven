package middleware

import (
	"auth-haven/internal/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ValidationMiddleware creates a middleware that validates request bodies
func ValidationMiddleware() gin.HandlerFunc {
	validator := validation.NewCustomValidator()
	
	return func(c *gin.Context) {
		// Skip validation for GET, DELETE, OPTIONS requests
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		
		// Get the content type
		contentType := c.GetHeader("Content-Type")
		
		// Only validate JSON requests
		if contentType != "application/json" {
			c.Next()
			return
		}
		
		// Store validator in context for handlers to use
		c.Set("validator", validator)
		c.Next()
	}
}

// ValidateStruct is a helper function for handlers to validate structs
func ValidateStruct(c *gin.Context, s interface{}) error {
	if validator, exists := c.Get("validator"); exists {
		if v, ok := validator.(*validation.CustomValidator); ok {
			return v.ValidateStruct(s)
		}
	}
	
	// Fallback to default validation
	return validation.NewCustomValidator().ValidateStruct(s)
}

// ValidateField is a helper function for handlers to validate individual fields
func ValidateField(c *gin.Context, field interface{}, tag string) error {
	if validator, exists := c.Get("validator"); exists {
		if v, ok := validator.(*validation.CustomValidator); ok {
			return v.ValidateVar(field, tag)
		}
	}
	
	// Fallback to default validation
	return validation.NewCustomValidator().ValidateVar(field, tag)
}

// ValidationErrorHandler returns a standardized validation error response
func ValidationErrorHandler(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "ValidationError",
		"message": "Invalid input data",
		"details": validation.GetValidationErrorMessage(err),
	})
}
