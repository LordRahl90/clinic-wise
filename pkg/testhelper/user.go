package testhelper

import (
	"clinic-wise/db/models"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

func CreateUser(db *gorm.DB, role models.UserRole) *models.User {
	user := &models.User{
		ID:        ulid.Make(),
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Email:     gofakeit.Email(),
		Password:  "password",
		Role:      role,
	}

	if err := db.Create(&user).Error; err != nil {
		return nil
	}
	return user
}
