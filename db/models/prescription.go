package models

import (
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type PrescriptionStatus string

const (
	ActivePrescription PrescriptionStatus = "active"
	Unavailable        PrescriptionStatus = "unavailable"
)

// Prescription once dispatched, they become unavailable
// audit log must track the lifecycle of this prescription eg. created, dispatched
type Prescription struct {
	ID             ulid.ULID          `json:"id" gorm:"size:50"`
	HospitalID     ulid.ULID          `json:"hospital_id"`
	DoctorID       ulid.ULID          `json:"doctor_id"`
	PatientID      ulid.ULID          `json:"patient_id"`
	AppointmentID  ulid.ULID          `json:"appointment_id"`
	ExpirationDate time.Time          `json:"expiration_date"`
	Details        string             `json:"details"`
	Status         PrescriptionStatus `json:"status"`
	gorm.Model
}
