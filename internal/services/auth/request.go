package auth

import (
	"clinic-wise/db/models"
	"strings"

	"github.com/oklog/ulid/v2"
)

type SignUpRequest struct {
	HospitalID string `json:"hospital_id" binding:"required"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type InviteUserRequest struct {
	HospitalID string          `json:"hospital_id" binding:"required"`
	FirstName  string          `json:"first_name" binding:"required"`
	LastName   string          `json:"last_name" binding:"required"`
	Email      string          `json:"email" binding:"required,email"`
	Role       models.UserRole `json:"role" binding:"required"`
}

type AcceptInviteRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

type ResetPasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func (r *SignUpRequest) ToModel() (*models.User, error) {
	hospitalID, err := ulid.ParseStrict(r.HospitalID)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:         ulid.Make(),
		HospitalID: hospitalID,
		FirstName:  strings.TrimSpace(r.FirstName),
		LastName:   strings.TrimSpace(r.LastName),
		Email:      normalizeEmail(r.Email),
		Role:       models.Patient,
		Accepted:   true,
	}, nil
}

func (r *InviteUserRequest) ToModel() (*models.User, error) {
	hospitalID, err := ulid.ParseStrict(r.HospitalID)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:         ulid.Make(),
		HospitalID: hospitalID,
		FirstName:  strings.TrimSpace(r.FirstName),
		LastName:   strings.TrimSpace(r.LastName),
		Email:      normalizeEmail(r.Email),
		Role:       r.Role,
		Accepted:   false,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
