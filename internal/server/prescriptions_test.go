package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"clinic-wise/db/models"
	prescriptionsservice "clinic-wise/internal/services/prescriptions"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

type prescriptionResponse struct {
	ID            string                    `json:"id"`
	AppointmentID string                    `json:"appointment_id"`
	Status        models.PrescriptionStatus `json:"status"`
}

func prescriptionFixture(t *testing.T, svr *Server) (*models.User, *models.User, *models.User, *models.Appointment, *prescriptionResponse) {
	t.Helper()
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	pharmacist := testhelper.CreateUser(db, models.Pharmacist)
	require.NotNil(t, pharmacist)

	hospitalID := ulid.Make()
	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  hospitalID,
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "prescription fixture",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	doctorToken, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
	require.NoError(t, err)

	payload, err := json.Marshal(prescriptionsservice.CreatePrescriptionRequest{
		HospitalID:     hospitalID.String(),
		PatientID:      patient.ID.String(),
		AppointmentID:  appointment.ID.String(),
		ExpirationDate: time.Now().Add(24 * time.Hour).UTC(),
		Details:        "Use once daily",
	})
	require.NoError(t, err)

	res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/prescriptions", doctorToken, string(payload))
	require.Equal(t, http.StatusOK, res.Code)

	var body prescriptionResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
	require.NotEmpty(t, body.ID)

	return doctor, patient, pharmacist, appointment, &body
}

func TestCreatePrescriptionRoute(t *testing.T) {
	svr := newAppointmentServer()
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
		Description: "create route",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	t.Run("requires doctor role", func(t *testing.T) {
		patientToken, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/prescriptions", patientToken, `{}`)
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("doctor can create", func(t *testing.T) {
		doctorToken, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		payload, err := json.Marshal(prescriptionsservice.CreatePrescriptionRequest{
			HospitalID:     hospitalID.String(),
			PatientID:      patient.ID.String(),
			AppointmentID:  appointment.ID.String(),
			ExpirationDate: time.Now().Add(12 * time.Hour).UTC(),
			Details:        "Take after meal",
		})
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/prescriptions", doctorToken, string(payload))
		require.Equal(t, http.StatusOK, res.Code)
	})
}

func TestDispatchPrescriptionRoute(t *testing.T) {
	svr := newAppointmentServer()
	doctor, _, pharmacist, _, prescription := prescriptionFixture(t, svr)

	t.Run("only pharmacists can dispatch", func(t *testing.T) {
		doctorToken, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/prescriptions/"+prescription.ID+"/dispatch", doctorToken, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("pharmacist can dispatch", func(t *testing.T) {
		pharmacistToken, err := testhelper.CreateToken(*pharmacist, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/prescriptions/"+prescription.ID+"/dispatch", pharmacistToken, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body prescriptionResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Equal(t, models.Unavailable, body.Status)
	})

	t.Run("unavailable prescriptions cannot be dispatched twice", func(t *testing.T) {
		freshDoctor := testhelper.CreateUser(db, models.Doctor)
		require.NotNil(t, freshDoctor)
		freshPatient := testhelper.CreateUser(db, models.Patient)
		require.NotNil(t, freshPatient)

		hospitalID := ulid.Make()
		appointment := &models.Appointment{
			ID:          ulid.Make(),
			HospitalID:  hospitalID,
			DoctorID:    freshDoctor.ID,
			PatientID:   freshPatient.ID,
			TimeslotID:  ulid.Make(),
			Description: "double dispatch route",
			Status:      models.AppointmentStatusActive,
		}
		require.NoError(t, db.Create(appointment).Error)

		doctorToken, err := testhelper.CreateToken(*freshDoctor, svr.config.SigningSecret)
		require.NoError(t, err)

		createPayload, err := json.Marshal(prescriptionsservice.CreatePrescriptionRequest{
			HospitalID:     hospitalID.String(),
			PatientID:      freshPatient.ID.String(),
			AppointmentID:  appointment.ID.String(),
			ExpirationDate: time.Now().Add(3 * time.Hour).UTC(),
			Details:        "Single dispatch only",
		})
		require.NoError(t, err)

		createRes := testhelper.NewRequest(t, svr.router, http.MethodPost, "/prescriptions", doctorToken, string(createPayload))
		require.Equal(t, http.StatusOK, createRes.Code)

		var created prescriptionResponse
		require.NoError(t, json.Unmarshal(createRes.Body.Bytes(), &created))

		pharmacistToken, err := testhelper.CreateToken(*pharmacist, svr.config.SigningSecret)
		require.NoError(t, err)

		firstDispatch := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/prescriptions/"+created.ID+"/dispatch", pharmacistToken, "")
		require.Equal(t, http.StatusOK, firstDispatch.Code)

		secondDispatch := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/prescriptions/"+created.ID+"/dispatch", pharmacistToken, "")
		require.Equal(t, http.StatusBadRequest, secondDispatch.Code)
	})

	t.Run("expired prescriptions cannot be dispatched", func(t *testing.T) {
		expiredDoctor := testhelper.CreateUser(db, models.Doctor)
		require.NotNil(t, expiredDoctor)
		expiredPatient := testhelper.CreateUser(db, models.Patient)
		require.NotNil(t, expiredPatient)

		hospitalID := ulid.Make()
		appointment := &models.Appointment{
			ID:          ulid.Make(),
			HospitalID:  hospitalID,
			DoctorID:    expiredDoctor.ID,
			PatientID:   expiredPatient.ID,
			TimeslotID:  ulid.Make(),
			Description: "expired dispatch route",
			Status:      models.AppointmentStatusActive,
		}
		require.NoError(t, db.Create(appointment).Error)

		doctorToken, err := testhelper.CreateToken(*expiredDoctor, svr.config.SigningSecret)
		require.NoError(t, err)

		createPayload, err := json.Marshal(prescriptionsservice.CreatePrescriptionRequest{
			HospitalID:     hospitalID.String(),
			PatientID:      expiredPatient.ID.String(),
			AppointmentID:  appointment.ID.String(),
			ExpirationDate: time.Now().Add(-1 * time.Hour).UTC(),
			Details:        "Expired medicine",
		})
		require.NoError(t, err)

		createRes := testhelper.NewRequest(t, svr.router, http.MethodPost, "/prescriptions", doctorToken, string(createPayload))
		require.Equal(t, http.StatusOK, createRes.Code)

		var created prescriptionResponse
		require.NoError(t, json.Unmarshal(createRes.Body.Bytes(), &created))

		pharmacistToken, err := testhelper.CreateToken(*pharmacist, svr.config.SigningSecret)
		require.NoError(t, err)
		dispatchRes := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/prescriptions/"+created.ID+"/dispatch", pharmacistToken, "")
		require.Equal(t, http.StatusBadRequest, dispatchRes.Code)
	})
}

func TestGetPrescriptionRoute(t *testing.T) {
	svr := newAppointmentServer()
	doctor, patient, pharmacist, _, prescription := prescriptionFixture(t, svr)

	t.Run("doctor can view details", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/prescriptions/"+prescription.ID, token, "")
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("patient can view details", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/prescriptions/"+prescription.ID, token, "")
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("pharmacist cannot view details", func(t *testing.T) {
		token, err := testhelper.CreateToken(*pharmacist, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/prescriptions/"+prescription.ID, token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})
}

func TestGetAppointmentPrescriptionsRoute(t *testing.T) {
	svr := newAppointmentServer()
	doctor, patient, pharmacist, appointment, _ := prescriptionFixture(t, svr)

	t.Run("doctor can list appointment prescriptions", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/prescriptions/appointment/"+appointment.ID.String(), token, "")
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("patient can list appointment prescriptions", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/prescriptions/appointment/"+appointment.ID.String(), token, "")
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("pharmacist cannot list appointment prescriptions", func(t *testing.T) {
		token, err := testhelper.CreateToken(*pharmacist, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/prescriptions/appointment/"+appointment.ID.String(), token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})
}
