package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"clinic-wise/db/models"
	authtoken "clinic-wise/pkg/auth"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

type sessionBody struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type userBody struct {
	ID string `json:"id"`
}

func TestAuthSignupRoute(t *testing.T) {
	svr := newAuthServer()

	t.Run("validation", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/signup", "", `{}`)
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("success", func(t *testing.T) {
		email := uniqueEmail("alice.patient")
		payload, err := json.Marshal(map[string]string{
			"hospital_id": ulid.Make().String(),
			"first_name":  "Alice",
			"last_name":   "Patient",
			"email":       email,
			"password":    "password123",
		})
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/signup", "", string(payload))
		require.Equal(t, http.StatusOK, res.Code)

		var body sessionBody
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.NotEmpty(t, body.AccessToken)
		require.NotEmpty(t, body.RefreshToken)
	})
}

func TestAuthSigninRoute(t *testing.T) {
	svr := newAuthServer()
	email := uniqueEmail("alice.patient")

	setupPayload, err := json.Marshal(map[string]string{
		"hospital_id": ulid.Make().String(),
		"first_name":  "Alice",
		"last_name":   "Patient",
		"email":       email,
		"password":    "password123",
	})
	require.NoError(t, err)
	setupRes := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/signup", "", string(setupPayload))
	require.Equal(t, http.StatusOK, setupRes.Code)

	t.Run("validation", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/signin", "", `{}`)
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("success", func(t *testing.T) {
		payload, err := json.Marshal(map[string]string{
			"email":    email,
			"password": "password123",
		})
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/signin", "", string(payload))
		require.Equal(t, http.StatusOK, res.Code)

		var body sessionBody
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.NotEmpty(t, body.AccessToken)
	})
}

func TestAuthInviteRoute(t *testing.T) {
	svr := newAuthServer()
	token := adminToken(t, svr)

	t.Run("validation", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/invite", token, `{}`)
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("rejects patients", func(t *testing.T) {
		email := uniqueEmail("bob.patient")
		payload, err := json.Marshal(map[string]string{
			"hospital_id": ulid.Make().String(),
			"first_name":  "Bob",
			"last_name":   "Patient",
			"email":       email,
			"role":        string(models.Patient),
		})
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/invite", token, string(payload))
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("success", func(t *testing.T) {
		email := uniqueEmail("bob.doctor")
		payload, err := json.Marshal(map[string]string{
			"hospital_id": ulid.Make().String(),
			"first_name":  "Bob",
			"last_name":   "Doctor",
			"email":       email,
			"role":        string(models.Doctor),
		})
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/invite", token, string(payload))
		require.Equal(t, http.StatusOK, res.Code)

		var body userBody
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.NotEmpty(t, body.ID)
	})
}

func TestAuthAcceptInviteRoute(t *testing.T) {
	svr := newAuthServer()

	invitee := &models.User{
		ID:         ulid.Make(),
		HospitalID: ulid.Make(),
		FirstName:  "Bob",
		LastName:   "Doctor",
		Email:      uniqueEmail("bob.accept"),
		Role:       models.Doctor,
		Accepted:   false,
	}
	require.NoError(t, db.Create(invitee).Error)

	t.Run("validation", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/invites/invalid/accept", "", `{}`)
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("success", func(t *testing.T) {
		payload, err := json.Marshal(map[string]string{"password": "password123"})
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/auth/invites/"+invitee.ID.String()+"/accept", "", string(payload))
		require.Equal(t, http.StatusOK, res.Code)

		var body sessionBody
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.NotEmpty(t, body.AccessToken)
	})
}

func TestAuthResetPasswordRoute(t *testing.T) {
	svr := newAuthServer()
	user := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, user)

	token, err := authtoken.GenerateToken(t.Context(), authtoken.TokenData{
		ID:            user.ID,
		Role:          user.Role,
		SessionID:     ulid.Make().String(),
		TokenCategory: authtoken.TokenCategoryAccess,
		Expiry:        int(time.Now().Add(15 * time.Minute).Unix()),
	}, svr.config.SigningSecret, 15*time.Minute)
	require.NoError(t, err)

	t.Run("validation", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/auth/password", token, `{}`)
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("success", func(t *testing.T) {
		payload, err := json.Marshal(map[string]string{
			"current_password": "password",
			"new_password":     "password456",
		})
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/auth/password", token, string(payload))
		require.Equal(t, http.StatusOK, res.Code)

		var body sessionBody
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.NotEmpty(t, body.AccessToken)
	})
}

func uniqueEmail(prefix string) string {
	return prefix + "+" + ulid.Make().String() + "@example.com"
}

func newAuthServer() *Server {
	return New(&Config{
		DB:            db,
		Port:          "8080",
		SigningSecret: "secret",
	})
}

func adminToken(t *testing.T, svr *Server) string {
	t.Helper()
	admin := testhelper.CreateUser(db, models.Admin)
	require.NotNil(t, admin)

	token, err := authtoken.GenerateToken(t.Context(), authtoken.TokenData{
		ID:            admin.ID,
		Role:          admin.Role,
		SessionID:     ulid.Make().String(),
		TokenCategory: authtoken.TokenCategoryAccess,
		Expiry:        int(time.Now().Add(15 * time.Minute).Unix()),
	}, svr.config.SigningSecret, 15*time.Minute)
	require.NoError(t, err)
	return token
}
