package auth

import (
	"context"
	"log"
	"os"
	"testing"

	"clinic-wise/db/migrator"
	"clinic-wise/db/models"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func TestMain(m *testing.M) {
	code := 1
	container, err := testhelper.GetMySQLContainer(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := container.Terminate(context.TODO()); err != nil {
			log.Fatal(err)
		}
		os.Exit(code)
	}()

	db = testhelper.SetupContainerTestDB(context.TODO(), container)
	if err := migrator.Migrate(db); err != nil {
		log.Fatal(err)
	}

	code = m.Run()
}

func TestService_SignUpAndSignIn(t *testing.T) {
	svc := New(db, "secret")
	hospitalID := ulid.Make().String()

	signUpRes, err := svc.SignUp(t.Context(), &SignUpRequest{
		HospitalID: hospitalID,
		FirstName:  "Jane",
		LastName:   "Doe",
		Email:      "jane.doe@example.com",
		Password:   "password123",
	})
	require.NoError(t, err)
	require.True(t, signUpRes.User.Accepted)
	require.NotEmpty(t, signUpRes.AccessToken)
	require.NotEmpty(t, signUpRes.RefreshToken)

	signInRes, err := svc.SignIn(t.Context(), &SignInRequest{
		Email:    "jane.doe@example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	require.Equal(t, signUpRes.User.Email, signInRes.User.Email)
	require.NotEmpty(t, signInRes.AccessToken)
}

func TestService_InviteAcceptAndResetPassword(t *testing.T) {
	svc := New(db, "secret")
	hospitalID := ulid.Make().String()

	inviteRes, err := svc.InviteUser(t.Context(), &InviteUserRequest{
		HospitalID: hospitalID,
		FirstName:  "John",
		LastName:   "Smith",
		Email:      "john.smith@example.com",
		Role:       models.Doctor,
	})
	require.NoError(t, err)
	require.False(t, inviteRes.Accepted)

	_, err = svc.SignIn(t.Context(), &SignInRequest{
		Email:    inviteRes.Email,
		Password: "password123",
	})
	require.Error(t, err)

	acceptedRes, err := svc.AcceptInvite(t.Context(), ulid.MustParse(inviteRes.ID), &AcceptInviteRequest{
		Password: "password123",
	})
	require.NoError(t, err)
	require.True(t, acceptedRes.User.Accepted)
	require.NotEmpty(t, acceptedRes.AccessToken)

	resetRes, err := svc.ResetPassword(t.Context(), ulid.MustParse(inviteRes.ID), &ResetPasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "password456",
	})
	require.NoError(t, err)
	require.NotEmpty(t, resetRes.AccessToken)

	_, err = svc.SignIn(t.Context(), &SignInRequest{
		Email:    inviteRes.Email,
		Password: "password123",
	})
	require.Error(t, err)

	signInRes, err := svc.SignIn(t.Context(), &SignInRequest{
		Email:    inviteRes.Email,
		Password: "password456",
	})
	require.NoError(t, err)
	require.Equal(t, inviteRes.Email, signInRes.User.Email)
}

func TestService_InviteRejectsPatients(t *testing.T) {
	svc := New(db, "secret")
	_, err := svc.InviteUser(t.Context(), &InviteUserRequest{
		HospitalID: ulid.Make().String(),
		FirstName:  "Pat",
		LastName:   "Ient",
		Email:      "patient.invite@example.com",
		Role:       models.Patient,
	})
	require.Error(t, err)
}
