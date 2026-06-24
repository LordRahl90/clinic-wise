package prescriptions

import (
	"clinic-wise/db/models"
	"time"
)

type Response struct {
	ID             string                    `json:"id"`
	HospitalID     string                    `json:"hospital_id"`
	DoctorID       string                    `json:"doctor_id"`
	PatientID      string                    `json:"patient_id"`
	AppointmentID  string                    `json:"appointment_id"`
	ExpirationDate time.Time                 `json:"expiration_date"`
	Details        string                    `json:"details"`
	Status         models.PrescriptionStatus `json:"status"`
}

func FromModel(m *models.Prescription) *Response {
	return &Response{
		ID:             m.ID.String(),
		HospitalID:     m.HospitalID.String(),
		DoctorID:       m.DoctorID.String(),
		PatientID:      m.PatientID.String(),
		AppointmentID:  m.AppointmentID.String(),
		ExpirationDate: m.ExpirationDate,
		Details:        m.Details,
		Status:         m.Status,
	}
}
