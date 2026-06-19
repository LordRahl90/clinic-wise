package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type AppointmentStatus string

const (
	AppointmentStatusActive    AppointmentStatus = "active"
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
	AppointmentStatusCompleted AppointmentStatus = "completed"
)

// Appointment represents an appointment between a doctor and a patient some assumptions here is that description carries
// all the information the patient wants to pass
type Appointment struct {
	ID          ulid.ULID         `json:"id" gorm:"size:50"`
	HospitalID  ulid.ULID         `json:"hospital_id"`
	DoctorID    ulid.ULID         `json:"doctor_id"`
	PatientID   ulid.ULID         `json:"patient_id"`
	TimeslotID  ulid.ULID         `json:"timeslot_id"`
	Description string            `json:"description"`
	Status      AppointmentStatus `json:"status"`
	gorm.Model

	Doctor  *User `json:"doctor" gorm:"foreignKey:DoctorID;references:ID"`
	Patient *User `json:"patient" gorm:"foreignKey:PatientID;references:ID"`
	//Prescriptions []Prescription `json:"prescriptions" gorm:"foreignKey:AppointmentID;references:ID"`
}

func (a *Appointment) BeforeCreate(_ *gorm.DB) (err error) {
	a.ID = ulid.Make()
	return
}
