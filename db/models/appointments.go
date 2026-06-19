package models

import "github.com/oklog/ulid/v2"

type Appointment struct {
	ID         ulid.ULID `json:"id" gorm:"size:50"`
	HospitalID ulid.ULID `json:"hospital_id"`
	DoctorID   ulid.ULID `json:"doctor_id"`
	UserID     ulid.ULID `json:"user_id"`
	TimeslotID ulid.ULID `json:"timeslot_id"`
}
