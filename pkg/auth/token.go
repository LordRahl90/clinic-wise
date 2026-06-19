package auth

import (
	"clinic-wise/db/models"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

type TokenCategory string

const (
	TokenCategoryUnknown TokenCategory = "unknown"
	TokenCategoryAccess  TokenCategory = "access"
	TokenCategoryRefresh TokenCategory = "refresh"
)

func (tc TokenCategory) String() string {
	return string(tc)
}

func TokenCategoryFromString(s string) TokenCategory {
	switch s {
	case TokenCategoryAccess.String():
		return TokenCategoryAccess
	case TokenCategoryRefresh.String():
		return TokenCategoryRefresh
	default:
		return TokenCategoryUnknown
	}
}

type TokenData struct {
	UserID        ulid.ULID       `json:"user_id"`
	Role          models.UserRole `json:"role"`
	SessionID     string          `json:"session_id"`
	TokenCategory TokenCategory   `json:"token_category"`
	Expiry        int             `json:"expiry"`
	jwt.RegisteredClaims
}

func GenerateToken(ctx context.Context, user TokenData, signingSecret string, duration time.Duration) (string, error) {
	slog.InfoContext(ctx, "Generating token", "user_id", user.ID)
	env := os.Getenv("ENVIRONMENT")
	// Create the token
	user.RegisteredClaims = jwt.RegisteredClaims{
		ID:        ulid.Make().String(),
		Subject:   user.UserID.String(),
		Issuer:    fmt.Sprintf("%s-%s", "clinic-wise", env),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	return token.SignedString([]byte(signingSecret))
}

// DecodeToken decodes the bearer token provided to extract basic users information
func DecodeToken(tokenString, signingSecret string) (*TokenData, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenData{}, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(signingSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	claims, ok := token.Claims.(*TokenData)
	if !ok {
		return nil, fmt.Errorf("error parsing token claims: %v", token.Claims)
	}

	return claims, nil
}
