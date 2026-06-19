package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"clinic-wise/db/models"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

type auditTrailResponse struct {
	ID            ulid.ULID `json:"id"`
	ActorID       ulid.ULID `json:"actor_id"`
	Action        string    `json:"action"`
	EntityType    string    `json:"entity_type"`
	EntityID      string    `json:"entity_id"`
	AppointmentID string    `json:"appointment_id"`
	Message       string    `json:"message"`
}

func newAuditServer() *Server {
	return New(&Config{
		DB:            db,
		Port:          "8080",
		SigningSecret: "secret",
	})
}

func TestListAppointmentAuditTrailRoute(t *testing.T) {
	svr := newAuditServer()
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	otherDoctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, otherDoctor)

	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  ulid.Make(),
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "audit trail route",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	require.NoError(t, db.Create(&models.AuditTrail{
		ID:            ulid.Make(),
		ActorID:       doctor.ID,
		ActorName:     "Doctor User",
		ActorRole:     string(models.Doctor),
		Action:        "note_created",
		EntityType:    "note",
		EntityID:      ulid.Make().String(),
		AppointmentID: appointment.ID.String(),
		Message:       "Doctor User added note to appointment " + appointment.ID.String(),
		Changes:       []byte("[]"),
	}).Error)

	t.Run("unauthorized returns 401", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/audit-trails/appointment/"+appointment.ID.String(), "", "")
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("doctor can view appointment audit", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/audit-trails/appointment/"+appointment.ID.String(), token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []auditTrailResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Len(t, body, 1)
		require.Equal(t, "note_created", body[0].Action)
		require.Equal(t, appointment.ID.String(), body[0].AppointmentID)
	})

	t.Run("patient can view appointment audit", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/audit-trails/appointment/"+appointment.ID.String(), token, "")
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("unrelated user cannot view appointment audit", func(t *testing.T) {
		token, err := testhelper.CreateToken(*otherDoctor, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/audit-trails/appointment/"+appointment.ID.String(), token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})
}

func TestListEntityAuditTrailRoute(t *testing.T) {
	svr := newAuditServer()
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	unrelated := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, unrelated)

	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  ulid.Make(),
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "entity route audit",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	noteID := ulid.Make()
	require.NoError(t, db.Create(&models.AuditTrail{
		ID:            ulid.Make(),
		ActorID:       doctor.ID,
		ActorName:     "Dr. House",
		ActorRole:     string(models.Doctor),
		Action:        "note_created",
		EntityType:    "note",
		EntityID:      noteID.String(),
		AppointmentID: appointment.ID.String(),
		Message:       "Dr. House added note to appointment " + appointment.ID.String(),
		Changes:       []byte("[]"),
	}).Error)

	url := "/audit-trails/entity/note/" + noteID.String()

	t.Run("unauthorized returns 401", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, url, "", "")
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("doctor linked to appointment can view", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, url, token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []auditTrailResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Len(t, body, 1)
		require.Equal(t, "note_created", body[0].Action)
		require.Contains(t, body[0].Message, "Dr. House")
	})

	t.Run("patient linked to appointment can view", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, url, token, "")
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("unrelated user cannot view", func(t *testing.T) {
		token, err := testhelper.CreateToken(*unrelated, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, url, token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("filters by action query param", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)

		res := testhelper.NewRequest(t, svr.router, http.MethodGet, url+"?action=note_created", token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []auditTrailResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Len(t, body, 1)

		res2 := testhelper.NewRequest(t, svr.router, http.MethodGet, url+"?action=prescription_created", token, "")
		require.Equal(t, http.StatusOK, res2.Code)

		var empty []auditTrailResponse
		require.NoError(t, json.Unmarshal(res2.Body.Bytes(), &empty))
		require.Len(t, empty, 0)
	})
}



