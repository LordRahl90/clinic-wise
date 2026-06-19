package timeslots

import (
	"clinic-wise/db/models"
	"clinic-wise/internal/services/audittrail"
	"context"
	"log/slog"

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
	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:    timeslot.UserID,
		Action:     "timeslot_created",
		EntityType: "timeslot",
		EntityID:   timeslot.ID.String(),
		Message:    "created timeslot on " + timeslot.Date,
		Changes: []audittrail.Change{
			{Field: "start_time", After: timeslot.StartTime},
			{Field: "end_time", After: timeslot.EndTime},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record timeslot create audit", "timeslot_id", timeslot.ID.String(), "error", err)
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
