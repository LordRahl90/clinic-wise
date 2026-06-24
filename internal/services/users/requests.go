package users

import (
	"clinic-wise/db/models"
	"strings"

	"github.com/oklog/ulid/v2"
)

type CreateUserRequest struct {
	HospitalID string          `json:"hospital_id" binding:"required"`
	FirstName  string          `json:"first_name" binding:"required"`
	LastName   string          `json:"last_name" binding:"required"`
	Email      string          `json:"email" binding:"required,email"`
	Role       models.UserRole `json:"role" binding:"required"`
}

func (r *CreateUserRequest) ToModel() (*models.User, error) {
	hospitalID, err := ulid.ParseStrict(r.HospitalID)
	if err != nil {
		return nil, err
	}

	return &models.User{
		HospitalID: hospitalID,
		FirstName:  strings.TrimSpace(r.FirstName),
		LastName:   strings.TrimSpace(r.LastName),
		Email:      strings.ToLower(strings.TrimSpace(r.Email)),
		Role:       r.Role,
	}, nil
}
