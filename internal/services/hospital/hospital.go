package hospital

import (
	"context"
	"database/sql"

	"clinic-wise/db/repositories"

	"github.com/oklog/ulid/v2"
)

type Service struct {
	queries *repositories.Queries
}

func New(db *sql.DB) *Service {
	return &Service{
		queries: repositories.New(db),
	}
}

func (s *Service) Create(ctx context.Context, req *CreateHospitalRequest) (*Response, error) {
	hospital := req.ToModel()
	err := s.queries.CreateHospital(ctx, hospital)
	if err != nil {
		return nil, err
	}

	return &Response{
		ID:   hospital.ID.String(),
		Name: hospital.Name.String,
	}, nil
}

func (s *Service) Find(ctx context.Context, id ulid.ULID) (*Response, error) {
	res, err := s.queries.GetHospital(ctx, id)
	if err != nil {
		return nil, err
	}

	return FromModel(res), nil
}
