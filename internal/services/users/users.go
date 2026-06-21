package users

import (
	"context"
	"database/sql"

	"clinic-wise/db/repositories"
)

type Service struct {
	queries *repositories.Queries
}

func New(db *sql.DB) *Service {
	return &Service{
		queries: repositories.New(db),
	}
}

func (s *Service) Create(ctx context.Context, req CreateUserRequest) (*Response, error) {
	user := req.ToModel()

	if err := s.queries.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return nil, nil
}
