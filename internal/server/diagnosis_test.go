package server

import (
	"clinic-wise/db/models"
	diagnosisservice "clinic-wise/internal/services/diagnosis"
	"clinic-wise/pkg/testhelper"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

type diagnosisResponse struct {
	ID            ulid.ULID `json:"id"`
	DoctorID      ulid.ULID `json:"doctor_id"`
	PatientID     ulid.ULID `json:"patient_id"`
	AppointmentID ulid.ULID `json:"appointment_id"`
	Diagnosis     string    `json:"diagnosis"`
	Details       string    `json:"details"`
	Dismissed     bool      `json:"dismissed"`
}

func newDiagnosisServer() *Server {
	return New(&Config{
		DB:            db,
		Port:          "8080",
		SigningSecret: "secret",
	})
}

func diagnosisFixture(t *testing.T, svr *Server) (*models.User, *models.User, *diagnosisResponse) {
	t.Helper()
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
		Description: "diagnosis fixture",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	doctorToken, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
	require.NoError(t, err)

	payload, err := json.Marshal(diagnosisservice.CreateDiagnosisRequest{
		HospitalID:    hospitalID,
		PatientID:     patient.ID,
		AppointmentID: appointment.ID,
		Diagnosis:     "Acute sinusitis",
		Details:       "Symptoms should resolve within two weeks",
	})
	require.NoError(t, err)

	res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/diagnoses", doctorToken, string(payload))
	require.Equal(t, http.StatusOK, res.Code)

	var body diagnosisResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
	require.NotEqual(t, ulid.ULID{}, body.ID)
	return doctor, patient, &body
}

func TestCreateDiagnosisRoute(t *testing.T) {
	svr := newDiagnosisServer()
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
		Description: "create diagnosis route",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	t.Run("unauthorized returns 401", func(t *testing.T) {
		payload, err := json.Marshal(diagnosisservice.CreateDiagnosisRequest{
			HospitalID:    hospitalID,
			PatientID:     patient.ID,
			AppointmentID: appointment.ID,
			Diagnosis:     "Allergy",
			Details:       "Avoid trigger",
		})
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/diagnoses", "", string(payload))
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("patients cannot create", func(t *testing.T) {
		patientToken, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/diagnoses", patientToken, `{}`)
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("doctor can create", func(t *testing.T) {
		doctorToken, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		payload, err := json.Marshal(diagnosisservice.CreateDiagnosisRequest{
			HospitalID:    hospitalID,
			PatientID:     patient.ID,
			AppointmentID: appointment.ID,
			Diagnosis:     "Tension headache",
			Details:       "Rest and hydration advised",
		})
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/diagnoses", doctorToken, string(payload))
		require.Equal(t, http.StatusOK, res.Code)
	})
}

func TestGetDiagnosisRoute(t *testing.T) {
	svr := newDiagnosisServer()
	doctor, patient, diagnosis := diagnosisFixture(t, svr)
	pharmacist := testhelper.CreateUser(db, models.Pharmacist)
	require.NotNil(t, pharmacist)

	t.Run("doctor can view details", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/diagnoses/"+diagnosis.ID.String(), token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body diagnosisResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Equal(t, diagnosis.ID, body.ID)
		require.Equal(t, diagnosis.Diagnosis, body.Diagnosis)
	})

	t.Run("patient can view details", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/diagnoses/"+diagnosis.ID.String(), token, "")
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("non-doctor non-patient cannot view", func(t *testing.T) {
		token, err := testhelper.CreateToken(*pharmacist, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/diagnoses/"+diagnosis.ID.String(), token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/diagnoses/not-valid", token, "")
		require.Equal(t, http.StatusBadRequest, res.Code)
	})
}

func TestDismissDiagnosisRoute(t *testing.T) {
	svr := newDiagnosisServer()
	doctor, patient, diagnosis := diagnosisFixture(t, svr)
	otherDoctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, otherDoctor)

	t.Run("unauthorized returns 401", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/diagnoses/"+diagnosis.ID.String()+"/dismiss", "", "")
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("patients cannot dismiss", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/diagnoses/"+diagnosis.ID.String()+"/dismiss", token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("only owning doctor can dismiss", func(t *testing.T) {
		token, err := testhelper.CreateToken(*otherDoctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/diagnoses/"+diagnosis.ID.String()+"/dismiss", token, "")
		require.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("doctor can dismiss", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/diagnoses/"+diagnosis.ID.String()+"/dismiss", token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body diagnosisResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.True(t, body.Dismissed)
	})
}

func TestListDiagnosisRoute(t *testing.T) {
	svr := newDiagnosisServer()
	doctor, patient, diagnosis := diagnosisFixture(t, svr)
	otherDoctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, otherDoctor)
	otherPatient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, otherPatient)
	pharmacist := testhelper.CreateUser(db, models.Pharmacist)
	require.NotNil(t, pharmacist)

	otherHospitalID := ulid.Make()
	otherAppt := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  otherHospitalID,
		DoctorID:    otherDoctor.ID,
		PatientID:   otherPatient.ID,
		TimeslotID:  ulid.Make(),
		Description: "unrelated diagnosis",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(otherAppt).Error)
	require.NoError(t, db.Create(&models.Diagnosis{
		ID:            ulid.Make(),
		HospitalID:    otherHospitalID,
		DoctorID:      otherDoctor.ID,
		PatientID:     otherPatient.ID,
		AppointmentID: otherAppt.ID,
		Diagnosis:     "Unrelated",
		Details:       "N/A",
	}).Error)

	t.Run("doctor sees only linked diagnoses", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/diagnoses", token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []diagnosisResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Len(t, body, 1)
		require.Equal(t, diagnosis.ID, body[0].ID)
		require.Equal(t, doctor.ID, body[0].DoctorID)
	})

	t.Run("patient sees only linked diagnoses", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/diagnoses", token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []diagnosisResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Len(t, body, 1)
		require.Equal(t, diagnosis.ID, body[0].ID)
		require.Equal(t, patient.ID, body[0].PatientID)
	})

	t.Run("non-doctor non-patient cannot list", func(t *testing.T) {
		token, err := testhelper.CreateToken(*pharmacist, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/diagnoses", token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})
}

