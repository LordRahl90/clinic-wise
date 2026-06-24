package diagnosis

import (
	"clinic-wise/db/models"
	"time"

	"github.com/oklog/ulid/v2"
)

type Response struct {
	ID            ulid.ULID `json:"id"`
	HospitalID    ulid.ULID `json:"hospital_id"`
	DoctorID      ulid.ULID `json:"doctor_id"`
	PatientID     ulid.ULID `json:"patient_id"`
	AppointmentID ulid.ULID `json:"appointment_id"`
	Diagnosis     string    `json:"diagnosis"`
	Details       string    `json:"details"`
	Dismissed     bool      `json:"dismissed"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func FromModel(m *models.Diagnosis) *Response {
	return &Response{
		ID:            m.ID,
		HospitalID:    m.HospitalID,
		DoctorID:      m.DoctorID,
		PatientID:     m.PatientID,
		AppointmentID: m.AppointmentID,
		Diagnosis:     m.Diagnosis,
		Details:       m.Details,
		Dismissed:     m.Dismissed,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}


