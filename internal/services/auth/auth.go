package auth

import (
	"clinic-wise/db/models"
	"clinic-wise/internal/services/audittrail"
	authtoken "clinic-wise/pkg/auth"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 30 * 24 * time.Hour
)

type Service struct {
	db            *gorm.DB
	signingSecret string
}

func New(db *gorm.DB, signingSecret string) *Service {
	return &Service{db: db, signingSecret: signingSecret}
}

func (s *Service) SignUp(ctx context.Context, req *SignUpRequest) (*SessionResponse, error) {
	user, err := req.ToModel()
	if err != nil {
		return nil, err
	}

	if err := s.ensureEmailAvailable(ctx, user.Email); err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user.Password = passwordHash

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:    user.ID,
		Action:     "user_signed_up",
		EntityType: "user",
		EntityID:   user.ID.String(),
		Message:    "created a new account",
		Changes: []audittrail.Change{
			{Field: "email", After: user.Email},
			{Field: "role", After: user.Role},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record signup audit", "user_id", user.ID.String(), "error", err)
	}

	return s.issueSession(ctx, user)
}

func (s *Service) SignIn(ctx context.Context, req *SignInRequest) (*SessionResponse, error) {
	user, err := s.findUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if !user.Accepted {
		return nil, fmt.Errorf("invitation has not been accepted")
	}

	if err := verifyPassword(user.Password, req.Password); err != nil {
		return nil, err
	}

	return s.issueSession(ctx, user)
}

func (s *Service) InviteUser(ctx context.Context, req *InviteUserRequest) (*UserResponse, error) {
	if req.Role == models.Patient {
		return nil, fmt.Errorf("patients can sign up directly")
	}

	user, err := req.ToModel()
	if err != nil {
		return nil, err
	}

	if err := s.ensureEmailAvailable(ctx, user.Email); err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:    ulid.ULID{},
		Action:     "user_invited",
		EntityType: "user",
		EntityID:   user.ID.String(),
		Message:    "invited " + user.Email + " as " + string(user.Role),
		Changes: []audittrail.Change{
			{Field: "email", After: user.Email},
			{Field: "role", After: user.Role},
			{Field: "accepted", After: user.Accepted},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record invite audit", "invited_user_id", user.ID.String(), "error", err)
	}

	res := UserFromModel(user)
	return &res, nil
}

func (s *Service) AcceptInvite(ctx context.Context, inviteID ulid.ULID, req *AcceptInviteRequest) (*SessionResponse, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", inviteID).First(&user).Error; err != nil {
		return nil, err
	}

	if user.Accepted {
		return nil, fmt.Errorf("invite has already been accepted")
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user.Password = passwordHash
	user.Accepted = true
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:    user.ID,
		Action:     "invite_accepted",
		EntityType: "user",
		EntityID:   user.ID.String(),
		Message:    "accepted invitation",
		Changes: []audittrail.Change{
			{Field: "accepted", Before: false, After: true},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record invite acceptance audit", "user_id", user.ID.String(), "error", err)
	}

	return s.issueSession(ctx, &user)
}

func (s *Service) ResetPassword(ctx context.Context, userID ulid.ULID, req *ResetPasswordRequest) (*SessionResponse, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	if !user.Accepted {
		return nil, fmt.Errorf("account has not been activated")
	}

	if err := verifyPassword(user.Password, req.CurrentPassword); err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	user.Password = passwordHash
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:    user.ID,
		Action:     "password_reset",
		EntityType: "user",
		EntityID:   user.ID.String(),
		Message:    "reset account password",
		Changes: []audittrail.Change{
			{Field: "password", Before: "***", After: "***"},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record password reset audit", "user_id", user.ID.String(), "error", err)
	}

	return s.issueSession(ctx, &user)
}

func (s *Service) findUserByEmail(ctx context.Context, email string) (*models.User, error) {
	normalizedEmail := normalizeEmail(email)
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) ensureEmailAvailable(ctx context.Context, email string) error {
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ?", normalizeEmail(email)).First(&user).Error
	if err == nil {
		return fmt.Errorf("email already exists")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (s *Service) issueSession(ctx context.Context, user *models.User) (*SessionResponse, error) {
	sessionID := ulid.Make().String()
	accessToken, err := authtoken.GenerateToken(ctx, authtoken.TokenData{
		ID:            user.ID,
		Role:          user.Role,
		SessionID:     sessionID,
		TokenCategory: authtoken.TokenCategoryAccess,
		Expiry:        int(time.Now().Add(accessTokenDuration).Unix()),
	}, s.signingSecret, accessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := authtoken.GenerateToken(ctx, authtoken.TokenData{
		ID:            user.ID,
		Role:          user.Role,
		SessionID:     sessionID,
		TokenCategory: authtoken.TokenCategoryRefresh,
		Expiry:        int(time.Now().Add(refreshTokenDuration).Unix()),
	}, s.signingSecret, refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return &SessionResponse{
		User:         UserFromModel(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func verifyPassword(storedPassword, providedPassword string) error {
	if storedPassword == "" {
		return fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(providedPassword)); err == nil {
		return nil
	}

	if strings.TrimSpace(storedPassword) == providedPassword {
		return nil
	}

	return fmt.Errorf("invalid credentials")
}
