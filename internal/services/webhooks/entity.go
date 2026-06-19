package webhooks

import (
	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
)

type RegisterWebhookRequest struct {
	HospitalID ulid.ULID `json:"hospital_id" binding:"required"`
	UserID     ulid.ULID `json:"user_id" binding:"required"`
	URL        string    `json:"url" binding:"required"`
}

func (r *RegisterWebhookRequest) ToModel() *models.Webhook {
	return &models.Webhook{
		ID:         ulid.Make(),
		HospitalID: r.HospitalID,
		PatientID:  r.UserID,
		URL:        r.URL,
	}
}

type Response struct {
	ID     ulid.ULID `json:"id"`
	UserID ulid.ULID `json:"user_id"`
	URL    string    `json:"url"`
}

func FromModel(m *models.Webhook) *Response {
	return &Response{
		ID:     m.ID,
		UserID: m.PatientID,
		URL:    m.URL,
	}
}
