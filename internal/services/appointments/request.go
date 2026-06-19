package appointments

import (
	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
)

type CreateAppointmentRequest struct {
	HospitalID  string `json:"hospital_id"`
	DoctorID    string `json:"doctor_id"`
	PatientID   string `json:"patient_id"`
	TimeslotID  string `json:"timeslot_id"`
	Description string `json:"description"`
}

// ToModel converts the CreateAppointmentRequest to a models.Appointment
// assumption here is that all new appointments are active by default
func (c *CreateAppointmentRequest) ToModel() (*models.Appointment, error) {
	hospitalID, err := ulid.ParseStrict(c.HospitalID)
	if err != nil {
		return nil, err
	}
	doctorID, err := ulid.ParseStrict(c.DoctorID)
	if err != nil {
		return nil, err
	}
	patientID, err := ulid.Parse(c.PatientID)
	if err != nil {
		return nil, err
	}
	timeslotID, err := ulid.Parse(c.TimeslotID)
	if err != nil {
		return nil, err
	}
	return &models.Appointment{
		HospitalID:  hospitalID,
		DoctorID:    doctorID,
		PatientID:   patientID,
		TimeslotID:  timeslotID,
		Description: c.Description,
		Status:      models.AppointmentStatusActive,
	}, nil
}
