package timeslots

import (
	"clinic-wise/db/models"
	"context"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, req *CreateTimeslotRequest) (*Response, error) {
	// validation would be done before this function is called
	// validation would also include overlapping timeslots for a specific user
	// this would also have a middleware that ensures that only doctors can create timeslots
	timeslot, err := req.ToModel()
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Create(timeslot).Error; err != nil {
		return nil, err
	}

	return ResponseFromModel(timeslot), nil
}

func (s *Service) FindByUser(ctx context.Context, userID ulid.ULID) ([]Response, error) {
	var timeslots []models.Timeslot
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&timeslots).Error; err != nil {
		return nil, err
	}
	result := make([]Response, 0, len(timeslots))
	for _, timeslot := range timeslots {
		result = append(result, *ResponseFromModel(&timeslot))
	}
	return result, nil
}
