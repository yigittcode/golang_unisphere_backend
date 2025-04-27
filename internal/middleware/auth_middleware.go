package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/auth"
)

// AuthMiddleware for authentication and authorization
type AuthMiddleware struct {
	jwtService *auth.JWTService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// JWTAuth middleware for JWT token validation
func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Authentication required")
			errorDetail = errorDetail.WithDetails("Authorization header missing")

			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
			return
		}

		// Extract token using the utility function
		tokenString, err := auth.ExtractBearerToken(authHeader)
		if err != nil {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Authentication required")
			errorDetail = errorDetail.WithDetails("Invalid token format")

			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
			return
		}

		// Validate and extract claims using the new method
		claims, err := m.jwtService.ValidateAndExtractClaims(tokenString)
		if err != nil {
			// Handle token errors in more detail
			statusCode := http.StatusUnauthorized
			errorCode := dto.ErrorCodeInvalidToken
			errorMessage := "Authentication failed"
			errorDetails := "Invalid token"

			if errors.Is(err, apperrors.ErrTokenExpired) {
				errorCode = dto.ErrorCodeExpiredToken
				errorDetails = "Token has expired"
			} else if errors.Is(err, apperrors.ErrInvalidFormat) {
				errorDetails = "Invalid token format"
			}

			errorDetail := dto.NewErrorDetail(errorCode, errorMessage)
			errorDetail = errorDetail.WithDetails(errorDetails)
			errorDetail = errorDetail.WithSeverity(dto.ErrorSeverityError)

			c.AbortWithStatusJSON(statusCode, dto.NewErrorResponse(errorDetail))
			return
		}

		// Add user information to context if token is valid
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roleType", claims.RoleType)

		c.Next()
	}
}

// RoleRequired middleware to check if user has required role
func (m *AuthMiddleware) RoleRequired(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure JWTAuth middleware has run first
		role, exists := c.Get("roleType")
		if !exists {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Authentication required")
			errorDetail = errorDetail.WithDetails("User role not found")

			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
			return
		}

		// Compare roles
		roleStr, ok := role.(string)
		if !ok || roleStr != requiredRole {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Access denied")
			errorDetail = errorDetail.WithDetails("You don't have sufficient permissions for this operation")
			errorDetail = errorDetail.WithSeverity(dto.ErrorSeverityError)

			c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse(errorDetail))
			return
		}

		c.Next()
	}
}
