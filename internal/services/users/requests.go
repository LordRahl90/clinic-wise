package users

import (
	"clinic-wise/db/repositories"
	"database/sql"
)

type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
}

func (c *CreateUserRequest) ToModel() repositories.CreateUserParams {
	return repositories.CreateUserParams{
		FirstName: sql.NullString{String: c.FirstName, Valid: true},
		LastName:  sql.NullString{String: c.LastName, Valid: true},
		Email:     sql.NullString{String: c.Email, Valid: true},
		Password:  sql.NullString{String: c.Password, Valid: true},
		Role:      sql.NullString{String: c.Role, Valid: true},
	}
}
