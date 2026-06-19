package server

import (
	authservice "clinic-wise/internal/services/auth"
	"context"
	"net/http"

	"clinic-wise/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type AuthService interface {
	SignUp(ctx context.Context, req *authservice.SignUpRequest) (*authservice.SessionResponse, error)
	SignIn(ctx context.Context, req *authservice.SignInRequest) (*authservice.SessionResponse, error)
	InviteUser(ctx context.Context, req *authservice.InviteUserRequest) (*authservice.UserResponse, error)
	AcceptInvite(ctx context.Context, inviteID ulid.ULID, req *authservice.AcceptInviteRequest) (*authservice.SessionResponse, error)
	ResetPassword(ctx context.Context, userID ulid.ULID, req *authservice.ResetPasswordRequest) (*authservice.SessionResponse, error)
}

func (s *Server) authRoutes() {
	authGroup := s.router.Group("/auth")
	{
		authGroup.POST("/signup", s.signUp)
		authGroup.POST("/signin", s.signIn)
		authGroup.POST("/invites/:id/accept", s.acceptInvite)
		authGroup.POST("/invite", s.authMiddleware.Middleware(), s.inviteUser)
		authGroup.PATCH("/password", s.authMiddleware.Middleware(), s.resetPassword)
	}
}

func (s *Server) signUp(c *gin.Context) {
	var req authservice.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.authService.SignUp(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) signIn(c *gin.Context) {
	var req authservice.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.authService.SignIn(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) inviteUser(c *gin.Context) {
	userInfo := currentUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req authservice.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.authService.InviteUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = userInfo
	c.JSON(http.StatusOK, res)
}

func (s *Server) acceptInvite(c *gin.Context) {
	inviteID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req authservice.AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.authService.AcceptInvite(c.Request.Context(), inviteID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) resetPassword(c *gin.Context) {
	userInfo := currentUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req authservice.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.authService.ResetPassword(c.Request.Context(), userInfo.ID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func currentUserInfo(c *gin.Context) *auth.TokenData {
	if userInfo, ok := c.Get("userInfo"); ok {
		if typed, ok := userInfo.(*auth.TokenData); ok {
			return typed
		}
	}
	return nil
}
