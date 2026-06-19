package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"clinic-wise/db/models"
	"clinic-wise/internal/services/hospital"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestHospitalStatsRoute(t *testing.T) {
	svr := New(&Config{
		DB:            db,
		Port:          "8080",
		SigningSecret: "secret",
	})

	hospitalID := ulid.Make()
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient1 := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient1)
	patient2 := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient2)
	nonAdmin := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, nonAdmin)

	appt1 := &models.Appointment{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient1.ID, TimeslotID: ulid.Make(), Description: "appt1", Status: models.AppointmentStatusActive}
	appt2 := &models.Appointment{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient2.ID, TimeslotID: ulid.Make(), Description: "appt2", Status: models.AppointmentStatusConfirmed}
	appt3 := &models.Appointment{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient1.ID, TimeslotID: ulid.Make(), Description: "appt3", Status: models.AppointmentStatusCancelled}
	require.NoError(t, db.Create(appt1).Error)
	require.NoError(t, db.Create(appt2).Error)
	require.NoError(t, db.Create(appt3).Error)

	require.NoError(t, db.Create(&models.Prescription{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient1.ID, AppointmentID: appt1.ID, ExpirationDate: time.Now().Add(24 * time.Hour), Details: "med1", Status: models.ActivePrescription}).Error)
	require.NoError(t, db.Create(&models.Prescription{ID: ulid.Make(), HospitalID: hospitalID, DoctorID: doctor.ID, PatientID: patient2.ID, AppointmentID: appt2.ID, ExpirationDate: time.Now().Add(24 * time.Hour), Details: "med2", Status: models.ActivePrescription}).Error)

	t.Run("unauthorized returns 401", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/hospitals/"+hospitalID.String()+"/stats", "", "")
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("non-admin returns 403", func(t *testing.T) {
		token, err := testhelper.CreateToken(*nonAdmin, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/hospitals/"+hospitalID.String()+"/stats", token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("admin can view stats", func(t *testing.T) {
		admin := testhelper.CreateUser(db, models.Admin)
		require.NotNil(t, admin)
		token, err := testhelper.CreateToken(*admin, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/hospitals/"+hospitalID.String()+"/stats", token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body hospital.StatsResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.EqualValues(t, 3, body.TotalAppointments)
		require.EqualValues(t, 2, body.ActivePatients)
		require.EqualValues(t, 2, body.PrescriptionCount)
	})
}
