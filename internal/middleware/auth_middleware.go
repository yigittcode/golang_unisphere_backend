package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/auth"
)

// AuthMiddleware for authentication and authorization
type AuthMiddleware struct {
	jwtService *auth.JWTService
	userRepo   *repositories.UserRepository
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(jwtService *auth.JWTService, userRepo *repositories.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

// JWTAuth middleware for JWT token validation
func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		var err error

		// Get Authorization header (standard method)
		authHeader := c.GetHeader("Authorization")

		// Check authorization query parameter if header is missing (for Swagger UI)
		if authHeader == "" {
			// Check if token is in query parameters (Swagger UI sometimes puts it here)
			if queryToken := c.Query("authorization"); queryToken != "" {
				authHeader = queryToken
			} else if queryToken := c.Query("Authorization"); queryToken != "" {
				authHeader = queryToken
			} else if queryToken := c.Query("token"); queryToken != "" {
				authHeader = queryToken
			}
		}

		// If still no token found, return unauthorized
		if authHeader == "" {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Authentication required")
			errorDetail = errorDetail.WithDetails("Authorization header missing")

			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
			return
		}

		// Extract token - Handle different token formats

		// Check if it's a raw JWT token (for Swagger UI convenience)
		if strings.Count(authHeader, ".") == 2 && !strings.HasPrefix(authHeader, "Bearer ") {
			// It looks like a raw JWT token, use it directly
			tokenString = authHeader
		} else {
			// Try normal extraction (requires Bearer prefix)
			tokenString, err = auth.ExtractBearerToken(authHeader)
			if err != nil {
				// One more attempt - maybe token is wrapped in quotes (happens with some clients)
				authHeader = strings.Trim(authHeader, "\"'")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				} else if strings.Count(authHeader, ".") == 2 {
					// It might still be a raw token
					tokenString = authHeader
				} else {
					errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Authentication required")
					errorDetail = errorDetail.WithDetails("Invalid token format")

					c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
					return
				}
			}
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

// EmailVerificationRequired middleware to check if user's email is verified
func (m *AuthMiddleware) EmailVerificationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by JWTAuth middleware)
		userID, exists := c.Get("userID")
		if !exists {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Authentication required")
			errorDetail = errorDetail.WithDetails("User information not found")
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
			return
		}

		// Convert to int64
		userIDInt, ok := userID.(int64)
		if !ok {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Internal server error")
			errorDetail = errorDetail.WithDetails("Invalid user ID format")
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
			return
		}

		// Check if email is verified
		verified, err := m.userRepo.IsEmailVerified(c.Request.Context(), userIDInt)
		if err != nil {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Internal server error")
			errorDetail = errorDetail.WithDetails("Failed to check email verification status")
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
			return
		}

		// If not verified, return forbidden
		if !verified {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeForbidden, "Email not verified")
			errorDetail = errorDetail.WithDetails("Please verify your email address before accessing this resource")
			c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse(errorDetail))
			return
		}

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
