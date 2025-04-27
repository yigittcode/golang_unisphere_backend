package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// JWTConfig defines JWT configuration settings
type JWTConfig struct {
	SecretKey       string
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration
	TokenIssuer     string
}

// JWTService handles JWT operations
type JWTService struct {
	config JWTConfig
}

// NewJWTService creates a new JWT service
func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{
		config: config,
	}
}

// Claims defines JWT token content
type Claims struct {
	UserID   int64  `json:"userId"`
	Email    string `json:"email"`
	RoleType string `json:"roleType"`
	jwt.RegisteredClaims
}

// GenerateTokenPair creates access and refresh token pair
func (s *JWTService) GenerateTokenPair(user *models.User) (accessToken, refreshToken string, expiresIn, refreshExpiresIn int, err error) {
	// Access token expiry
	accessTokenExpiry := time.Now().Add(s.config.AccessTokenExp)

	// Create claims
	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		RoleType: string(user.RoleType),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.TokenIssuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			ID:        uuid.New().String(),
		},
	}

	// Create access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("failed to create access token: %w", err)
	}

	// Create refresh token (simple UUID)
	refreshToken = uuid.New().String()

	// Token durations in seconds
	expiresIn = int(s.config.AccessTokenExp.Seconds())
	refreshExpiresIn = int(s.config.RefreshTokenExp.Seconds())

	return accessToken, refreshToken, expiresIn, refreshExpiresIn, nil
}

// ValidateToken validates a token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		// Token expired
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrTokenExpired
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Get claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, apperrors.ErrTokenInvalid
}

// GetRefreshTokenExpiry returns refresh token expiry time
func (s *JWTService) GetRefreshTokenExpiry() time.Time {
	return time.Now().Add(s.config.RefreshTokenExp)
}

// ExtractBearerToken extracts the token from the Authorization header
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", apperrors.ErrInvalidFormat
	}

	// Check if the header starts with "Bearer " (required)
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	// If in Swagger UI, the user might paste just the token without "Bearer " prefix
	// To be more user-friendly, check if the header looks like a JWT token (three dot-separated parts)
	if strings.Count(authHeader, ".") == 2 {
		// It looks like a JWT token, so return it directly
		return authHeader, nil
	}

	// Otherwise, reject it as an invalid format
	return "", fmt.Errorf("%w: authorization header must start with 'Bearer '", apperrors.ErrInvalidFormat)
}

// ValidateAndExtractClaims validates and extracts claims from a token string
func (s *JWTService) ValidateAndExtractClaims(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, apperrors.ErrTokenInvalid
	}

	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Additional validation if needed
	if claims.UserID <= 0 || claims.Email == "" {
		return nil, apperrors.ErrTokenInvalid
	}

	return claims, nil
}
