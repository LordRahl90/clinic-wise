package hospital

import "clinic-wise/db/models"

type CreateHospitalResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func FromModel(m *models.Hospital) *CreateHospitalResponse {
	return &CreateHospitalResponse{
		ID:   m.ID.String(),
		Name: m.Name,
	}
}

type StatsResponse struct {
	TotalAppointments int64 `json:"total_appointments"`
	ActivePatients    int64 `json:"active_patients"`
	PrescriptionCount int64 `json:"prescription_count"`
}
