package timeslots

import "clinic-wise/db/models"

type Response struct {
	ID         string `json:"id"`
	HospitalID string `json:"hospital_id"`
	UserID     string `json:"user_id"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func ResponseFromModel(m *models.Timeslot) *Response {
	return &Response{
		ID: m.ID.String(),
	}
}
