package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Note struct {
	ID            ulid.ULID `json:"id" gorm:"size:50"`
	HospitalID    ulid.ULID `json:"hospital_id"`
	AppointmentID ulid.ULID `json:"appointment_id"`
	DoctorID      ulid.ULID `json:"doctor_id"`
	PatientID     ulid.ULID `json:"patient_id"`
	Content       string    `json:"content"`

	gorm.Model
}
