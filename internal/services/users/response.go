package users

import (
	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
)

type Response struct {
	HospitalID ulid.ULID `json:"hospital_id"`
	ID         ulid.ULID `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
}

func FromModel(m *models.User) *Response {
	return &Response{
		HospitalID: m.HospitalID,
		ID:         m.ID,
		FirstName:  m.FirstName,
		LastName:   m.LastName,
		Email:      m.Email,
	}
}
