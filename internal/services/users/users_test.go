package users

import (
	"context"
	"log"
	"os"
	"testing"

	"clinic-wise/db/migrator"
	"clinic-wise/db/models"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var db *gorm.DB

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

func TestService_Create_NewUser(t *testing.T) {
	svc := NewService(db)
	res, err := svc.Create(t.Context(), &CreateUserRequest{
		HospitalID: ulid.Make().String(),
		FirstName:  "Alice",
		LastName:   "Smith",
		Email:      "alice.smith@example.com",
		Role:       models.Doctor,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "alice.smith@example.com", res.Email)
	assert.Equal(t, "Alice", res.FirstName)
	assert.Equal(t, "Smith", res.LastName)
	assert.NotEmpty(t, res.ID)
}

func TestService_Create_ExistingUserReturnedByEmail(t *testing.T) {
	svc := NewService(db)
	first, err := svc.Create(t.Context(), &CreateUserRequest{
		HospitalID: ulid.Make().String(),
		FirstName:  "Bob",
		LastName:   "Jones",
		Email:      "bob.jones@example.com",
		Role:       models.Patient,
	})
	require.NoError(t, err)
	require.NotNil(t, first)

	// Same email, different casing — should return the existing user unchanged.
	second, err := svc.Create(t.Context(), &CreateUserRequest{
		HospitalID: ulid.Make().String(),
		FirstName:  "Robert",
		LastName:   "Jones",
		Email:      "BOB.JONES@example.com",
		Role:       models.Admin,
	})
	require.NoError(t, err)
	require.NotNil(t, second)

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, first.Email, second.Email)
}

func TestService_Create_InvalidHospitalID(t *testing.T) {
	svc := NewService(db)
	_, err := svc.Create(t.Context(), &CreateUserRequest{
		HospitalID: "not-a-valid-ulid",
		FirstName:  "Carol",
		LastName:   "White",
		Email:      "carol.white@example.com",
		Role:       models.Pharmacist,
	})
	require.Error(t, err)
}

func TestService_FindByEmail_Found(t *testing.T) {
	svc := NewService(db)
	_, err := svc.Create(t.Context(), &CreateUserRequest{
		HospitalID: ulid.Make().String(),
		FirstName:  "Diana",
		LastName:   "Prince",
		Email:      "diana.prince@example.com",
		Role:       models.Doctor,
	})
	require.NoError(t, err)

	res, err := svc.FindByEmail(t.Context(), "diana.prince@example.com")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "diana.prince@example.com", res.Email)
	assert.Equal(t, "Diana", res.FirstName)
	assert.Equal(t, "Prince", res.LastName)
}

func TestService_FindByEmail_CaseInsensitive(t *testing.T) {
	svc := NewService(db)
	_, err := svc.Create(t.Context(), &CreateUserRequest{
		HospitalID: ulid.Make().String(),
		FirstName:  "Eve",
		LastName:   "Adams",
		Email:      "eve.adams@example.com",
		Role:       models.Pharmacist,
	})
	require.NoError(t, err)

	res, err := svc.FindByEmail(t.Context(), "EVE.ADAMS@EXAMPLE.COM")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "eve.adams@example.com", res.Email)
}

func TestService_FindByEmail_NotFound(t *testing.T) {
	svc := NewService(db)
	_, err := svc.FindByEmail(t.Context(), "nonexistent@example.com")
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
