package transcriber

import "github.com/oklog/ulid/v2"

type StartTranscribingRequest struct {
	AppointmentID ulid.ULID `json:"appointment_id"`
}

type TranscriptionResponse struct {
}
