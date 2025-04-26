package middleware

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/yigit/unisphere/internal/app/models/dto"
)

var validate = validator.New()

// ValidationMiddleware handles request validation
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// ValidateRequest validates a request body against the provided model
func ValidateRequest(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(obj); err != nil {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid request format")
			errorDetail = errorDetail.WithDetails(err.Error())
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
			c.Abort()
			return
		}

		// Use reflect to get a pointer to the actual value if needed
		value := reflect.ValueOf(obj)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		// Only validate if the object implements validation
		if err := validate.Struct(value.Interface()); err != nil {
			errorDetail := dto.HandleValidationError(err)
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
			c.Abort()
			return
		}

		// Set the validated object in the context
		c.Set("validatedBody", obj)
		c.Next()
	}
}

// formatValidationError creates a human-readable validation error message
func formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "min":
		return e.Field() + " must be at least " + e.Param()
	case "max":
		return e.Field() + " must be at most " + e.Param()
	case "email":
		return e.Field() + " must be a valid email address"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	default:
		return e.Field() + " validation failed: " + e.Tag()
	}
}
