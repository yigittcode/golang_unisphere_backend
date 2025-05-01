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

// ValidateJsonRequest validates a JSON request body against the provided model
// Content-Type: application/json için kullanılır
func ValidateJsonRequest(obj interface{}) gin.HandlerFunc {
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

// ValidateFormDataRequest validates a multipart/form-data request against the provided model
// Content-Type: multipart/form-data için kullanılır (dosya yükleme desteği)
func ValidateFormDataRequest(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ShouldBind otomatik olarak form verilerini (form-data, x-www-form-urlencoded veya multipart/form-data) hedef yapıya bağlar
		if err := c.ShouldBind(obj); err != nil {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid form data")
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

		// Validate the struct after binding
		if err := validate.Struct(value.Interface()); err != nil {
			errorDetail := dto.HandleValidationError(err)
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
			c.Abort()
			return
		}

		// Set the validated object in the context
		c.Set("validatedFormData", obj)
		c.Next()
	}
}

// Eski metot isimleri için uyumluluk sağlayacak yönlendirici fonksiyonlar
// Geriye dönük uyumluluk için eski fonksiyon isimlerini koruyoruz
// Ancak bunlar yeni fonksiyonlara yönlendirir

// ValidateRequest uses ValidateJsonRequest (backward compatibility)
func ValidateRequest(obj interface{}) gin.HandlerFunc {
	return ValidateJsonRequest(obj)
}

// ValidateFormRequest uses ValidateFormDataRequest (backward compatibility)
func ValidateFormRequest(obj interface{}) gin.HandlerFunc {
	return ValidateFormDataRequest(obj)
}
