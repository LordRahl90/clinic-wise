package hospital

import (
	"context"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, req *CreateHospitalRequest) (*CreateHospitalResponse, error) {
	hospital := req.ToModel()
	if err := s.db.WithContext(ctx).Create(hospital).Error; err != nil {
		return nil, err
	}

	return &CreateHospitalResponse{
		ID: hospital.ID.String(),
	}, nil
}
