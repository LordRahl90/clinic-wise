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
