package notes

import (
	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
)

type CreateNoteRequest struct {
	HospitalID    ulid.ULID `json:"hospital_id"`
	PatientID     ulid.ULID `json:"patient_id"`
	DoctorID      ulid.ULID `json:"doctor_id"`
	AppointmentID ulid.ULID `json:"appointment_id"`
	Content       string    `json:"content"`
}

type CreateNoteStreamingRequest struct {
}

func (r *CreateNoteRequest) ToModel() *models.Note {
	return &models.Note{
		ID:            ulid.Make(),
		HospitalID:    ulid.Make(),
		PatientID:     r.PatientID,
		DoctorID:      r.DoctorID,
		AppointmentID: r.AppointmentID,
		Content:       r.Content,
	}
}

type StartDictationRequest struct {
	HospitalID    ulid.ULID `json:"hospital_id"`
	DoctorID      ulid.ULID `json:"-"`
	AppointmentID ulid.ULID `json:"appointment_id"`
}

type NoteFeeback struct {
	AppointmentID ulid.ULID `json:"appointmentId"`
	Sequence      int       `json:"sequence"`
	Text          string    `json:"text"`
	IsFinal       bool      `json:"isFinal"`
}
