package appointments

import (
	"clinic-wise/db/models"
	"context"
	"log"
	"os"
	"testing"

	"clinic-wise/db/migrator"
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

func TestService_Create(t *testing.T) {
	svc := New(db)
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	req := &CreateAppointmentRequest{
		PatientID:   patient.ID.String(),
		DoctorID:    doctor.ID.String(),
		TimeslotID:  ulid.Make().String(),
		HospitalID:  ulid.Make().String(),
		Description: "Test appointment",
	}
	res, err := svc.Create(t.Context(), req)
	require.NoError(t, err)
	require.NotEmpty(t, res.ID)
}
