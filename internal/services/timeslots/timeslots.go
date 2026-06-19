package timeslots

import (
	"clinic-wise/db/models"
	"context"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, req *CreateTimeslotRequest) error {
	timeslot, err := req.ToModel()
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Create(timeslot).Error
}

func (s *Service) FindByUser(ctx context.Context, userID string) ([]models.Timeslot, error) {
	var timeslots []models.Timeslot
	return timeslots, s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&timeslots).Error
}
