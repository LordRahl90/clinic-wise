package diagnosis

import (
	"clinic-wise/db/migrator"
	"clinic-wise/db/models"
	"clinic-wise/pkg/testhelper"
	"context"
	"log"
	"os"
	"testing"

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
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)

	hospitalID := ulid.Make()
	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  hospitalID,
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "diagnosis create",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	svc := New(db)
	created, err := svc.Create(t.Context(), &CreateDiagnosisRequest{
		HospitalID:    hospitalID,
		DoctorID:      doctor.ID,
		PatientID:     patient.ID,
		AppointmentID: appointment.ID,
		Diagnosis:     "Acute bronchitis",
		Details:       "Rest and hydration recommended",
	})
	require.NoError(t, err)
	require.NotEqual(t, ulid.ULID{}, created.ID)
	require.Equal(t, doctor.ID, created.DoctorID)
	require.Equal(t, patient.ID, created.PatientID)
	require.Equal(t, appointment.ID, created.AppointmentID)

	var stored models.Diagnosis
	require.NoError(t, db.Where("id = ?", created.ID).First(&stored).Error)
	require.Equal(t, created.Diagnosis, stored.Diagnosis)
	require.Equal(t, created.Details, stored.Details)
}

func TestService_Find(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	otherDoctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, otherDoctor)

	hospitalID := ulid.Make()
	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  hospitalID,
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "diagnosis lookup",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	svc := New(db)
	created, err := svc.Create(t.Context(), &CreateDiagnosisRequest{
		HospitalID:    hospitalID,
		DoctorID:      doctor.ID,
		PatientID:     patient.ID,
		AppointmentID: appointment.ID,
		Diagnosis:     "Seasonal allergy",
		Details:       "Avoid pollen exposure",
	})
	require.NoError(t, err)

	found, err := svc.Find(t.Context(), doctor.ID, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
	require.Equal(t, created.PatientID, found.PatientID)

	patientFound, err := svc.Find(t.Context(), patient.ID, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, patientFound.ID)

	_, err = svc.Find(t.Context(), otherDoctor.ID, created.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestService_Dismiss(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	otherDoctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, otherDoctor)

	hospitalID := ulid.Make()
	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  hospitalID,
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "diagnosis dismiss",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	svc := New(db)
	created, err := svc.Create(t.Context(), &CreateDiagnosisRequest{
		HospitalID:    hospitalID,
		DoctorID:      doctor.ID,
		PatientID:     patient.ID,
		AppointmentID: appointment.ID,
		Diagnosis:     "Migraine",
		Details:       "Use prescribed pain management",
	})
	require.NoError(t, err)
	require.False(t, created.Dismissed)

	_, err = svc.Dismiss(t.Context(), otherDoctor.ID, created.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	dismissed, err := svc.Dismiss(t.Context(), doctor.ID, created.ID)
	require.NoError(t, err)
	require.True(t, dismissed.Dismissed)

	var stored models.Diagnosis
	require.NoError(t, db.Where("id = ?", created.ID).First(&stored).Error)
	require.True(t, stored.Dismissed)
}

func TestService_FindByUser(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	otherDoctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, otherDoctor)
	otherPatient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, otherPatient)

	hospitalID := ulid.Make()
	svc := New(db)

	createAppointment := func(doctorID, patientID ulid.ULID, description string) *models.Appointment {
		appt := &models.Appointment{
			ID:          ulid.Make(),
			HospitalID:  hospitalID,
			DoctorID:    doctorID,
			PatientID:   patientID,
			TimeslotID:  ulid.Make(),
			Description: description,
			Status:      models.AppointmentStatusActive,
		}
		require.NoError(t, db.Create(appt).Error)
		return appt
	}

	appt1 := createAppointment(doctor.ID, patient.ID, "doctor and patient linked")
	_, err := svc.Create(t.Context(), &CreateDiagnosisRequest{
		HospitalID:    hospitalID,
		DoctorID:      doctor.ID,
		PatientID:     patient.ID,
		AppointmentID: appt1.ID,
		Diagnosis:     "Flu",
		Details:       "Hydration",
	})
	require.NoError(t, err)

	appt2 := createAppointment(otherDoctor.ID, patient.ID, "same patient only")
	_, err = svc.Create(t.Context(), &CreateDiagnosisRequest{
		HospitalID:    hospitalID,
		DoctorID:      otherDoctor.ID,
		PatientID:     patient.ID,
		AppointmentID: appt2.ID,
		Diagnosis:     "Allergy",
		Details:       "Antihistamine",
	})
	require.NoError(t, err)

	appt3 := createAppointment(otherDoctor.ID, otherPatient.ID, "unrelated")
	_, err = svc.Create(t.Context(), &CreateDiagnosisRequest{
		HospitalID:    hospitalID,
		DoctorID:      otherDoctor.ID,
		PatientID:     otherPatient.ID,
		AppointmentID: appt3.ID,
		Diagnosis:     "Unrelated",
		Details:       "N/A",
	})
	require.NoError(t, err)

	doctorList, err := svc.FindByUser(t.Context(), doctor.ID)
	require.NoError(t, err)
	require.Len(t, doctorList, 1)
	require.Equal(t, doctor.ID, doctorList[0].DoctorID)

	patientList, err := svc.FindByUser(t.Context(), patient.ID)
	require.NoError(t, err)
	require.Len(t, patientList, 2)
	for _, item := range patientList {
		require.Equal(t, patient.ID, item.PatientID)
	}
}



