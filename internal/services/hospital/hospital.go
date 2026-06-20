package hospital

import (
	"clinic-wise/db/repositories"
	"context"
	"database/sql"
)

type Service struct {
	queries *repositories.Queries
}

func New(db *sql.DB) *Service {
	return &Service{
		queries: repositories.New(db),
	}
}

func (s *Service) Create(ctx context.Context, req *CreateHospitalRequest) (*CreateHospitalResponse, error) {
	hospital := req.ToModel()
	err := s.queries.CreateHospital(ctx, hospital)
	if err != nil {
		return nil, err
	}

	return &CreateHospitalResponse{
		ID:   hospital.ID.String(),
		Name: hospital.Name.String,
	}, nil
}
