package hospital

import (
	"clinic-wise/db/repositories"
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"
)

type CreateHospitalRequest struct {
	UserID ulid.ULID `json:"user_id"`
	Name   string    `json:"name"`
}

func (r *CreateHospitalRequest) ToModel() repositories.CreateHospitalParams {
	return repositories.CreateHospitalParams{
		ID:        ulid.Make(),
		Name:      sql.NullString{String: r.Name, Valid: true},
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}
}
