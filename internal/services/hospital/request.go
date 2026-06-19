package hospital

import "clinic-wise/db/models"

type CreateHospitalRequest struct {
	Name string `json:"name"`
}

func (r *CreateHospitalRequest) ToModel() *models.Hospital {
	return &models.Hospital{
		Name: r.Name,
	}
}
