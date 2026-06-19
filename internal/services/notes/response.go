package notes

import (
	"clinic-wise/db/models"
	"time"

	"github.com/oklog/ulid/v2"
)

type Response struct {
	ID            ulid.ULID `json:"id"`
	AppointmentID ulid.ULID `json:"appointment_id"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func FromModel(m *models.Note) *Response {
	return &Response{
		ID:            m.ID,
		AppointmentID: m.AppointmentID,
		Content:       m.Content,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

type StreamingResponse struct {
	ID string `json:"id"`
}

type StreamingResponseData struct {
	AppointmentID string `json:"appointment_id"`
	Sequence      int    `json:"sequence"`
	Text          string `json:"text"`
	IsFinal       bool   `json:"is_final"`
}
