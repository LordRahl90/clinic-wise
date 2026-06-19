package hospital

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

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

func TestCreate(t *testing.T) {
	svc := New(db)
	req := &CreateHospitalRequest{
		Name: "Test Hospital",
	}
	res, err := svc.Create(t.Context(), req)
	require.NoError(t, err)
	require.NotEmpty(t, res.ID)
}

func TestStats(t *testing.T) {
	svc := New(db)
	hospitalID := ulid.Make()
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient1 := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient1)
	patient2 := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient2)
	patient3 := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient3)

	activeAppointment1 := &models.Appointment{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient1.ID, TimeslotID: ulid.Make(), Description: "a1", Status: models.AppointmentStatusActive}
	confirmedAppointment := &models.Appointment{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient2.ID, TimeslotID: ulid.Make(), Description: "a2", Status: models.AppointmentStatusConfirmed}
	activeAppointment2 := &models.Appointment{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient1.ID, TimeslotID: ulid.Make(), Description: "a3", Status: models.AppointmentStatusActive}
	cancelledAppointment := &models.Appointment{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient3.ID, TimeslotID: ulid.Make(), Description: "a4", Status: models.AppointmentStatusCancelled}
	require.NoError(t, db.Create(activeAppointment1).Error)
	require.NoError(t, db.Create(confirmedAppointment).Error)
	require.NoError(t, db.Create(activeAppointment2).Error)
	require.NoError(t, db.Create(cancelledAppointment).Error)

	require.NoError(t, db.Create(&models.Prescription{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient1.ID, AppointmentID: activeAppointment1.ID, ExpirationDate: time.Now().Add(24 * time.Hour), Details: "med1", Status: models.ActivePrescription}).Error)
	require.NoError(t, db.Create(&models.Prescription{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient2.ID, AppointmentID: confirmedAppointment.ID, ExpirationDate: time.Now().Add(24 * time.Hour), Details: "med2", Status: models.ActivePrescription}).Error)
	require.NoError(t, db.Create(&models.Prescription{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient1.ID, AppointmentID: activeAppointment2.ID, ExpirationDate: time.Now().Add(24 * time.Hour), Details: "med3", Status: models.ActivePrescription}).Error)

	res, err := svc.Stats(t.Context(), hospitalID)
	require.NoError(t, err)
	require.EqualValues(t, 4, res.TotalAppointments)
	require.EqualValues(t, 2, res.ActivePatients)
	require.EqualValues(t, 3, res.PrescriptionCount)
}
