package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"clinic-wise/db/models"
	"clinic-wise/internal/services/appointments"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

// ── POST /appointments ──────────────────────────────────────────────────────

func TestCreateAppointmentRoute(t *testing.T) {
	svr := newAppointmentServer()

	t.Run("no auth returns 401", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/appointments", "", `{}`)
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("validation - missing fields returns 500 from service", func(t *testing.T) {
		doctor := testhelper.CreateUser(db, models.Doctor)
		require.NotNil(t, doctor)
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		// empty body ⇒ all IDs blank ⇒ ulid.ParseStrict fails ⇒ service error
		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/appointments", token, `{}`)
		require.Equal(t, http.StatusInternalServerError, res.Code)
	})

	t.Run("success", func(t *testing.T) {
		doctor := testhelper.CreateUser(db, models.Doctor)
		require.NotNil(t, doctor)
		patient := testhelper.CreateUser(db, models.Patient)
		require.NotNil(t, patient)

		payload, err := json.Marshal(appointments.CreateAppointmentRequest{
			HospitalID:  ulid.Make().String(),
			DoctorID:    doctor.ID.String(),
			PatientID:   patient.ID.String(),
			TimeslotID:  ulid.Make().String(),
			Description: "Annual check-up",
		})
		require.NoError(t, err)

		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/appointments", token, string(payload))
		require.Equal(t, http.StatusOK, res.Code)

		var body appointments.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.NotEmpty(t, body.ID)
	})
}

// ── GET /appointments/:id ───────────────────────────────────────────────────

func TestFindAppointmentRoute(t *testing.T) {
	svr := newAppointmentServer()
	hospitalID := ulid.Make()
	doctor, patient, appt := appointmentFixture(t, svr, hospitalID)

	t.Run("no auth returns 401", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/"+appt.ID, "", "")
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("invalid id format returns 400", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/not-a-ulid", token, "")
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("unrelated user is denied", func(t *testing.T) {
		outsider := testhelper.CreateUser(db, models.Doctor)
		require.NotNil(t, outsider)
		token, err := testhelper.CreateToken(*outsider, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/"+appt.ID, token, "")
		require.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("doctor can access their appointment", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/"+appt.ID, token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body appointments.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Equal(t, appt.ID, body.ID)
	})

	t.Run("patient can access their appointment", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/"+appt.ID, token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body appointments.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Equal(t, appt.ID, body.ID)
	})
}

// ── GET /appointments/user ──────────────────────────────────────────────────

func TestFindAppointmentsByUserRoute(t *testing.T) {
	svr := newAppointmentServer()
	hospitalID := ulid.Make()
	doctor, patient, appt := appointmentFixture(t, svr, hospitalID)

	t.Run("no auth returns 401", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/user?page=1&limit=10", "", "")
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("missing page returns 400", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/user?limit=10", token, "")
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("missing limit returns 400", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/user?page=1", token, "")
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("doctor sees their appointments", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/user?page=1&limit=10", token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []appointments.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.NotEmpty(t, body)
		require.True(t, containsID(body, appt.ID), "doctor should see their appointment in the list")
	})

	t.Run("patient sees their appointments", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/user?page=1&limit=10", token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []appointments.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.NotEmpty(t, body)
		require.True(t, containsID(body, appt.ID), "patient should see their appointment in the list")
	})

	t.Run("unrelated user sees no shared appointments", func(t *testing.T) {
		outsider := testhelper.CreateUser(db, models.Patient)
		require.NotNil(t, outsider)
		token, err := testhelper.CreateToken(*outsider, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/appointments/user?page=1&limit=10", token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []appointments.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.False(t, containsID(body, appt.ID), "outsider should not see the fixture appointment")
	})
}

// ── GET /hospitals/:hospitalId/appointments ────────────────────────────────

func TestFindAppointmentsRoute(t *testing.T) {
	svr := newAppointmentServer()
	hospitalID := ulid.Make()
	_, _, appt := appointmentFixture(t, svr, hospitalID)

	admin := testhelper.CreateUser(db, models.Admin)
	require.NotNil(t, admin)
	token, err := testhelper.CreateToken(*admin, svr.config.SigningSecret)
	require.NoError(t, err)

	t.Run("no auth returns 401", func(t *testing.T) {
		url := "/hospitals/" + hospitalID.String() + "/appointments"
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, url, "", "")
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("invalid hospital_id returns 400", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/hospitals/not-valid/appointments", token, "")
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("returns appointments for the given hospital", func(t *testing.T) {
		url := "/hospitals/" + hospitalID.String() + "/appointments"
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, url, token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []appointments.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.NotEmpty(t, body)
		require.True(t, containsID(body, appt.ID), "appointment should appear under its hospital")
	})

	t.Run("different hospital_id returns empty list", func(t *testing.T) {
		url := "/hospitals/" + ulid.Make().String() + "/appointments"
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, url, token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []appointments.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.False(t, containsID(body, appt.ID), "appointment should not appear under a different hospital")
	})
}

func TestCompleteAppointmentRoute(t *testing.T) {
	svr := newAppointmentServer()
	hospitalID := ulid.Make()
	doctor, patient, appt := appointmentFixture(t, svr, hospitalID)

	t.Run("no auth returns 401", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/appointments/"+appt.ID+"/complete", "", "")
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("patient cannot complete", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/appointments/"+appt.ID+"/complete", token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("doctor can complete", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/appointments/"+appt.ID+"/complete", token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body appointments.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Equal(t, appt.ID, body.ID)

		var stored models.Appointment
		apptID, err := ulid.ParseStrict(appt.ID)
		require.NoError(t, err)
		require.NoError(t, db.Where("id = ?", apptID).First(&stored).Error)
		require.Equal(t, models.AppointmentStatusCompleted, stored.Status)
	})
}

// newAppointmentServer returns a test server with a consistent signing secret.
func newAppointmentServer() *Server {
	return New(&Config{
		DB:            db,
		Port:          "8080",
		SigningSecret: "secret",
	})
}

// appointmentFixture creates a doctor, patient, and a persisted appointment between them.
func appointmentFixture(t *testing.T, svr *Server, hospitalID ulid.ULID) (*models.User, *models.User, *appointments.Response) {
	t.Helper()
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)

	payload, err := json.Marshal(appointments.CreateAppointmentRequest{
		HospitalID:  hospitalID.String(),
		DoctorID:    doctor.ID.String(),
		PatientID:   patient.ID.String(),
		TimeslotID:  ulid.Make().String(),
		Description: "Fixture appointment",
	})
	require.NoError(t, err)

	token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
	require.NoError(t, err)

	res := testhelper.NewRequest(t, svr.router, http.MethodPost, "/appointments", token, string(payload))
	require.Equal(t, http.StatusOK, res.Code)

	var body appointments.Response
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
	require.NotEmpty(t, body.ID)

	return doctor, patient, &body
}

// containsID reports whether a response slice contains the given appointment ID.
func containsID(list []appointments.Response, id string) bool {
	for _, r := range list {
		if r.ID == id {
			return true
		}
	}
	return false
}
