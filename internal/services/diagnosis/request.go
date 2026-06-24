package diagnosis

import (
	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
)

type CreateDiagnosisRequest struct {
	HospitalID    ulid.ULID `json:"hospital_id" binding:"required"`
	PatientID     ulid.ULID `json:"patient_id" binding:"required"`
	AppointmentID ulid.ULID `json:"appointment_id" binding:"required"`
	Diagnosis     string    `json:"diagnosis" binding:"required"`
	Details       string    `json:"details" binding:"required"`

	// Filled from the authenticated doctor in the server layer.
	DoctorID ulid.ULID `json:"-"`
}

func (r *CreateDiagnosisRequest) ToModel() *models.Diagnosis {
	return &models.Diagnosis{
		ID:            ulid.Make(),
		HospitalID:    r.HospitalID,
		DoctorID:      r.DoctorID,
		PatientID:     r.PatientID,
		AppointmentID: r.AppointmentID,
		Diagnosis:     r.Diagnosis,
		Details:       r.Details,
	}
}

