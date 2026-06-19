package hospital

import (
	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
)

type CreateHospitalRequest struct {
	UserID ulid.ULID `json:"user_id"`
	Name   string    `json:"name"`
}

func (r *CreateHospitalRequest) ToModel() *models.Hospital {
	return &models.Hospital{
		ID:   ulid.Make(),
		Name: r.Name,
	}
}
