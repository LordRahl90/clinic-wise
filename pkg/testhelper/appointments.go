package testhelper

import (
	"clinic-wise/db/models"
	"log"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func CreateAppointment(t *testing.T, db *gorm.DB) *models.Appointment {
	t.Helper()
	doctor := CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := CreateUser(db, models.Patient)
	require.NotNil(t, patient)

	appt := &models.Appointment{
		ID:        ulid.Make(),
		DoctorID:  doctor.ID,
		PatientID: patient.ID,
	}
	if err := db.Create(appt).Error; err != nil {
		log.Fatal(err)
	}
	return appt
}
