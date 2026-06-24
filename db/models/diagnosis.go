package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Diagnosis struct {
	ID            ulid.ULID `json:"id" gorm:"size:50"`
	HospitalID    ulid.ULID `json:"hospital_id"`
	DoctorID      ulid.ULID `json:"doctor_id"`
	PatientID     ulid.ULID `json:"patient_id"`
	AppointmentID ulid.ULID `json:"appointment_id"`
	Diagnosis     string    `json:"diagnosis"`
	Details       string    `json:"details"`
	Dismissed     bool      `json:"dismissed" gorm:"default:false"`

	gorm.Model
}

func (d *Diagnosis) BeforeCreate(_ *gorm.DB) (err error) {
	if d.ID == (ulid.ULID{}) {
		d.ID = ulid.Make()
	}
	return
}

