package seeder

import (
	"context"

	"clinic-wise/internal/services/users"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

const (
	adminEmail = "alugbin.abiodun@gmail.com"
)

func SeedBaseUser(ctx context.Context, dbb *gorm.DB) error {
	userService := users.NewService(dbb)
	_, err := userService.Create(ctx, &users.CreateUserRequest{
		HospitalID: ulid.Make().String(),
		FirstName:  "John",
		LastName:   "Doe",
		Email:      adminEmail,
		Password:   "admin123",
		Role:       "admin",
	})
	return err
}
