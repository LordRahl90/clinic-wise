package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"clinic-wise/db/models"
	notesservice "clinic-wise/internal/services/notes"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

type noteResponse struct {
	ID            ulid.ULID `json:"id"`
	AppointmentID ulid.ULID `json:"appointment_id"`
	Content       string    `json:"content"`
}

func noteFixture(t *testing.T) (*Server, *models.User, *models.User, *models.Appointment, *models.Note) {
	t.Helper()
	svr := newAppointmentServer()
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)

	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  ulid.Make(),
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "notes fixture",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	note := &models.Note{
		ID:            ulid.Make(),
		HospitalID:    appointment.HospitalID,
		AppointmentID: appointment.ID,
		DoctorID:      doctor.ID,
		PatientID:     patient.ID,
		Content:       "Initial note",
	}
	require.NoError(t, db.Create(note).Error)

	return svr, doctor, patient, appointment, note
}

func TestUpdateNoteRoute(t *testing.T) {
	svr, doctor, patient, _, note := noteFixture(t)

	payload, err := json.Marshal(map[string]string{
		"note_id": note.ID.String(),
		"content": " | Follow up",
	})
	require.NoError(t, err)

	t.Run("unauthorized returns 401", func(t *testing.T) {
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/notes", "", string(payload))
		require.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("patient cannot update note", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/notes", token, string(payload))
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("invalid note_id returns 400", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/notes", token, `{"note_id":"bad-id","content":"x"}`)
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("other doctor cannot update", func(t *testing.T) {
		otherDoctor := testhelper.CreateUser(db, models.Doctor)
		require.NotNil(t, otherDoctor)
		token, err := testhelper.CreateToken(*otherDoctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/notes", token, string(payload))
		require.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("owning doctor can update", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodPatch, "/notes", token, string(payload))
		require.Equal(t, http.StatusOK, res.Code)

		var stored models.Note
		require.NoError(t, db.Where("id = ?", note.ID).First(&stored).Error)
		require.Equal(t, "Initial note | Follow up", stored.Content)
	})
}

func TestGetNoteRoute(t *testing.T) {
	svr, doctor, patient, _, note := noteFixture(t)

	t.Run("doctor can get note", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/notes/"+note.ID.String(), token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body noteResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Equal(t, note.ID, body.ID)
		require.Equal(t, note.Content, body.Content)
	})

	t.Run("patient can get note", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/notes/"+note.ID.String(), token, "")
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("pharmacist cannot get note", func(t *testing.T) {
		pharmacist := testhelper.CreateUser(db, models.Pharmacist)
		require.NotNil(t, pharmacist)
		token, err := testhelper.CreateToken(*pharmacist, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/notes/"+note.ID.String(), token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("unrelated doctor gets 404", func(t *testing.T) {
		otherDoctor := testhelper.CreateUser(db, models.Doctor)
		require.NotNil(t, otherDoctor)
		token, err := testhelper.CreateToken(*otherDoctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/notes/"+note.ID.String(), token, "")
		require.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/notes/not-valid", token, "")
		require.Equal(t, http.StatusBadRequest, res.Code)
	})
}

func TestGetAppointmentNotesRoute(t *testing.T) {
	svr, doctor, patient, appointment, note := noteFixture(t)

	secondNote := &models.Note{
		ID:            ulid.Make(),
		HospitalID:    appointment.HospitalID,
		AppointmentID: appointment.ID,
		DoctorID:      doctor.ID,
		PatientID:     patient.ID,
		Content:       "Second note",
	}
	require.NoError(t, db.Create(secondNote).Error)

	unrelatedAppointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  ulid.Make(),
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "unrelated appointment",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(unrelatedAppointment).Error)
	require.NoError(t, db.Create(&models.Note{
		ID:            ulid.Make(),
		HospitalID:    unrelatedAppointment.HospitalID,
		AppointmentID: unrelatedAppointment.ID,
		DoctorID:      doctor.ID,
		PatientID:     patient.ID,
		Content:       "Other appointment note",
	}).Error)

	path := "/notes/appointment/" + appointment.ID.String()

	t.Run("doctor can list appointment notes", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, path, token, "")
		require.Equal(t, http.StatusOK, res.Code)

		var body []notesservice.Response
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		require.Len(t, body, 2)
		returned := map[ulid.ULID]bool{}
		for _, n := range body {
			returned[n.ID] = true
		}
		require.True(t, returned[note.ID])
		require.True(t, returned[secondNote.ID])
	})

	t.Run("patient can list appointment notes", func(t *testing.T) {
		token, err := testhelper.CreateToken(*patient, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, path, token, "")
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("pharmacist cannot list appointment notes", func(t *testing.T) {
		pharmacist := testhelper.CreateUser(db, models.Pharmacist)
		require.NotNil(t, pharmacist)
		token, err := testhelper.CreateToken(*pharmacist, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, path, token, "")
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("unrelated doctor gets 404", func(t *testing.T) {
		otherDoctor := testhelper.CreateUser(db, models.Doctor)
		require.NotNil(t, otherDoctor)
		token, err := testhelper.CreateToken(*otherDoctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, path, token, "")
		require.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		token, err := testhelper.CreateToken(*doctor, svr.config.SigningSecret)
		require.NoError(t, err)
		res := testhelper.NewRequest(t, svr.router, http.MethodGet, "/notes/appointment/not-valid", token, "")
		require.Equal(t, http.StatusBadRequest, res.Code)
	})
}

