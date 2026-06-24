package prescriptions

import (
	"time"

	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
)

type CreatePrescriptionRequest struct {
	HospitalID     string    `json:"hospital_id" binding:"required"`
	PatientID      string    `json:"patient_id" binding:"required"`
	AppointmentID  string    `json:"appointment_id" binding:"required"`
	ExpirationDate time.Time `json:"expiration_date" binding:"required"`
	Details        string    `json:"details" binding:"required"`

	// Filled from the authenticated user in the server layer.
	DoctorID string `json:"doctor_id"`
}

func (r *CreatePrescriptionRequest) ToModel() (*models.Prescription, error) {
	hospitalID, err := ulid.ParseStrict(r.HospitalID)
	if err != nil {
		return nil, err
	}
	doctorID, err := ulid.ParseStrict(r.DoctorID)
	if err != nil {
		return nil, err
	}
	patientID, err := ulid.ParseStrict(r.PatientID)
	if err != nil {
		return nil, err
	}
	appointmentID, err := ulid.ParseStrict(r.AppointmentID)
	if err != nil {
		return nil, err
	}

	return &models.Prescription{
		ID:             ulid.Make(),
		HospitalID:     hospitalID,
		DoctorID:       doctorID,
		PatientID:      patientID,
		AppointmentID:  appointmentID,
		ExpirationDate: r.ExpirationDate,
		Details:        r.Details,
		Status:         models.ActivePrescription,
	}, nil
}
