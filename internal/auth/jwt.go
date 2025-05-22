// internal/auth/jwt.go
package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// Claims includes our custom fields plus the standard claims.
type Claims struct {
	UserID    uint   `json:"user_id"`
	TenantID  uint   `json:"tenant_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.StandardClaims
}

// GenerateAccessToken creates an access token (short-lived).
func GenerateAccessToken(userID, tenantID uint, role, secret string) (string, error) {
	claims := Claims{
		UserID:    userID,
		TenantID:  tenantID,
		Role:      role,
		TokenType: "access",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(150 * time.Minute).Unix(), // Access token valid for 15 minutes
			IssuedAt:  time.Now().Unix(),
			Issuer:    "myapp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a refresh token (longer-lived).
func GenerateRefreshToken(userID, tenantID uint, role, secret string) (string, error) {
	claims := Claims{
		UserID:    userID,
		TenantID:  tenantID,
		Role:      role,
		TokenType: "refresh",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(), // Refresh token valid for 7 days
			IssuedAt:  time.Now().Unix(),
			Issuer:    "myapp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken validates and returns claims for either token type.
func ParseToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// ParseRefreshToken makes sure the given token is a refresh token.
func ParseRefreshToken(tokenStr, secret string) (*Claims, error) {
	claims, err := ParseToken(tokenStr, secret)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("provided token is not a refresh token")
	}
	return claims, nil
}

// AuthMiddleware validates the access token from the Authorization header and
// stores claims in the request context.
type contextKey string

const (
	// ContextKeyClaims is used to store claims in the context.
	ContextKeyClaims = contextKey("claims")
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	log.Println("AuthMiddleware initialized with secret:", secret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]
			claims, err := ParseToken(tokenStr, secret)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Store the claims in context for later handlers.
			ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
