package testhelper

import (
	"context"
	"time"

	"clinic-wise/db/models"
	"clinic-wise/pkg/auth"

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
		Accepted:  true,
	}

	if err := db.Create(&user).Error; err != nil {
		return nil
	}
	return user
}

func CreateToken(user models.User, signingSecret string) (string, error) {
	tokenData := auth.TokenData{
		ID:            user.ID,
		Role:          user.Role,
		TokenCategory: auth.TokenCategoryAccess,
		Expiry:        int(time.Now().Add(time.Hour * 24 * 7).Unix()),
	}
	return auth.GenerateToken(context.Background(), tokenData, signingSecret, time.Hour*24)
}
