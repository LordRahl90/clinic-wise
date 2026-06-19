package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"clinic-wise/pkg/auth"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	signingSecret string
}

func NewAuthMiddleware(signingSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		signingSecret: signingSecret,
	}
}

func (am *AuthMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()

		userInfo, err := ExtractUserInfo(c, am.signingSecret)
		if err != nil || userInfo == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if userInfo.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			c.Abort()
		}

		if userInfo.TokenCategory == auth.TokenCategoryRefresh || userInfo.TokenCategory.String() == "" {
			slog.WarnContext(c.Request.Context(), "token is for refresh", "user_id", userInfo.ID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			c.Abort()
		}

		slog.InfoContext(c.Request.Context(), "user authenticated", "user", userInfo.ID, "path", path)
		c.Set("userInfo", userInfo)

		// Read and restore the request body so downstream handlers still see it.
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
	}
}

//func (am *AuthMiddleware) RefreshMiddleware() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		path := c.FullPath()
//
//		headers := c.Request.Header
//		authHeader, ok := headers["Authorization"]
//		if !ok {
//			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
//			c.Abort()
//			return
//		}
//
//		userInfo, err := ExtractUserInfo(c, am.signingSecret)
//		if err != nil || userInfo == nil {
//			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
//			c.Abort()
//			return
//		}
//
//		if userInfo.TokenCategory == auth.TokenCategoryAccess || userInfo.TokenCategory.String() == "" {
//			slog.WarnContext(c.Request.Context(), "refresh token is for access", "user_id", userInfo.UserID)
//			c.JSON(http.StatusBadGateway, gin.H{"error": "invalid token"})
//			c.Abort()
//		}
//
//		slog.InfoContext(c.Request.Context(), "user authenticated", "user", userInfo.UserID, "path", path)
//		c.Set("userInfo", userInfo)
//		c.Next()
//	}
//}

func ExtractUserInfo(c *gin.Context, signingSecret string) (*auth.TokenData, error) {
	authHeader, ok := c.Request.Header["Authorization"]
	if !ok {
		return nil, fmt.Errorf("authorization header not provided")
	}
	ctx := c.Request.Context()

	authData := strings.Split(authHeader[0], " ")
	if len(authData) != 2 {
		slog.WarnContext(ctx, "invalid token format provided", "token", authData)
		return nil, fmt.Errorf("invalid token format provided")
	}
	authToken := authData[1]
	userInfo, err := auth.DecodeToken(authToken, signingSecret)
	if err != nil || userInfo == nil {
		slog.ErrorContext(ctx, "failed to decode token", "error", err)
		return nil, fmt.Errorf("failed to decode token %w", err)
	}

	return userInfo, nil
}
