package server

import (
	"clinic-wise/db/models"
	"clinic-wise/internal/services/appointments"
	"clinic-wise/pkg/testhelper"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestCreateAppointment(t *testing.T) {
	svr := New(&Config{
		DB:            db,
		Port:          "8080",
		SigningSecret: "secret",
	})

	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)

	req := appointments.CreateAppointmentRequest{
		PatientID:   patient.ID.String(),
		DoctorID:    doctor.ID.String(),
		TimeslotID:  ulid.Make().String(),
		HospitalID:  ulid.Make().String(),
		Description: "Test appointment",
	}

	b, err := json.Marshal(req)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	res := testhelper.NewRequest(t, svr.router, "POST", "/appointments", "", string(b))
	require.Equal(t, http.StatusOK, res.Code)
}
