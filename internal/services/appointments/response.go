package appointments

import "clinic-wise/db/models"

type Response struct {
	ID string `json:"id"`
}

func ResponseFromModel(m *models.Appointment) *Response {
	appt := &Response{
		ID: m.ID.String(),
	}

	// we might need to load more information including the doctor's details, notes and prescriptions
	// but for now we will keep it simple and just return the appointment ID

	return appt
}
