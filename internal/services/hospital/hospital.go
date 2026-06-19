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

func (s *Service) Create(ctx context.Context, req *CreateHospitalRequest) error {
	hospital := req.ToModel()
	return s.db.WithContext(ctx).Create(hospital).Error
}
