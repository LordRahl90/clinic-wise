package auth

import "clinic-wise/db/models"

type UserResponse struct {
	ID         string          `json:"id"`
	HospitalID string          `json:"hospital_id"`
	FirstName  string          `json:"first_name"`
	LastName   string          `json:"last_name"`
	Email      string          `json:"email"`
	Role       models.UserRole `json:"role"`
	Accepted   bool            `json:"accepted"`
}

type SessionResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

func UserFromModel(m *models.User) UserResponse {
	return UserResponse{
		ID:         m.ID.String(),
		HospitalID: m.HospitalID.String(),
		FirstName:  m.FirstName,
		LastName:   m.LastName,
		Email:      m.Email,
		Role:       m.Role,
		Accepted:   m.Accepted,
	}
}
