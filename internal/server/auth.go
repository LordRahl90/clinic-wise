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

// signUp godoc
//
//	@Summary		Sign up a new patient
//	@Description	Registers a new patient account and returns a session with access/refresh tokens.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		swaggerSignUpRequest	true	"Sign-up payload"
//	@Success		200		{object}	swaggerSessionResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/auth/signup [post]
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

// signIn godoc
//
//	@Summary		Sign in
//	@Description	Authenticates a user and returns a session with access/refresh tokens.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		swaggerSignInRequest	true	"Sign-in payload"
//	@Success		200		{object}	swaggerSessionResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/auth/signin [post]
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

// inviteUser godoc
//
//	@Summary		Invite a user
//	@Description	Sends an invitation to a non-patient user (doctor, pharmacist, admin). Requires authentication.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		swaggerInviteUserRequest	true	"Invite payload"
//	@Success		200		{object}	swaggerUserResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/auth/invite [post]
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

// acceptInvite godoc
//
//	@Summary		Accept an invitation
//	@Description	Accepts a pending invite by setting the user's password, activating the account.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Invite ID (ULID)"
//	@Param			body	body		swaggerAcceptInviteRequest	true	"Accept invite payload"
//	@Success		200		{object}	swaggerSessionResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/auth/invites/{id}/accept [post]
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

// resetPassword godoc
//
//	@Summary		Reset password
//	@Description	Changes the authenticated user's password.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		swaggerResetPasswordRequest	true	"Reset password payload"
//	@Success		200		{object}	swaggerSessionResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/auth/password [patch]
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
