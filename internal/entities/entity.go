package entities

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type EventType string

const (
	AppointmentCreated  EventType = "appointment_created"
	AppointmentUpdated  EventType = "appointment_updated"
	AppointmentDeleted  EventType = "appointment_deleted"
	NoteCreated         EventType = "note_created"
	NoteUpdated         EventType = "note_updated"
	NoteDeleted         EventType = "note_deleted"
	PrescriptionCreated EventType = "prescription_created"
	PrescriptionUpdated EventType = "prescription_updated"
	PrescriptionDeleted EventType = "prescription_deleted"
)

const (
	NoteEventType         = "note"
	PrescriptionEventType = "prescription"
)

type Event struct {
	EventID       string      `json:"eventId"`
	EventType     EventType   `json:"eventType"`
	Timestamp     time.Time   `json:"timestamp"`
	AppointmentID ulid.ULID   `json:"appointmentId"`
	PatientID     ulid.ULID   `json:"patientId"`
	Payload       interface{} `json:"data"`
}

type NotePayload struct {
	NoteID   string `json:"noteId"`
	NoteText string `json:"noteText"`
}

type PrescriptionPayload struct {
	PrescriptionID string    `json:"prescriptionId"`
	Medication     string    `json:"medication"`
	ExpiresAt      time.Time `json:"expiresAt"`
}
