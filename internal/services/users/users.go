package users

import (
	"clinic-wise/db/models"
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// FindByEmail returns the user with the given email, or gorm.ErrRecordNotFound if none exists.
func (s *Service) FindByEmail(ctx context.Context, email string) (*Response, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Where("email = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return FromModel(&user), nil
}

// Create returns an existing user if one with the same email already exists,
// otherwise creates a new user and returns it.
func (s *Service) Create(ctx context.Context, req *CreateUserRequest) (*Response, error) {
	user, err := req.ToModel()
	if err != nil {
		return nil, err
	}

	existing, err := s.FindByEmail(ctx, user.Email)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return FromModel(user), nil
}
