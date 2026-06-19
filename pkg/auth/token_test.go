package auth

import (
	"os"
	"testing"
	"time"

	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	signingSecret = "secret"
)

var (
	db *gorm.DB
)

func TestMain(m *testing.M) {
	code := 1
	defer func() {
		os.Exit(code)
	}()

	code = m.Run()
}

func TestGenerateToken(t *testing.T) {
	user := TokenData{
		UserID: ulid.Make(),
		Role:   models.Doctor,
	}

	token, err := GenerateToken(t.Context(), user, signingSecret, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	t.Log(token)

	res, err := DecodeToken(token, signingSecret)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.UserID, user.UserID)
}
