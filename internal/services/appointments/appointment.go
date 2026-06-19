package appointments

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

func (s *Service) Create(ctx context.Context, req *CreateAppointmentRequest) (*Response, error) {
	m, err := req.ToModel()
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}

	return ResponseFromModel(m), nil
}

// Find finds an appointment by ID
// this can only be retrieved by the patient or doctor so a userID is required
func (s *Service) Find(ctx context.Context, userID, id ulid.ULID) (*Response, error) {
	var m models.Appointment
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}

	if m.DoctorID != userID && m.PatientID != userID {
		// user is not authorized to view this appointment
		return nil, gorm.ErrRecordNotFound
	}

	return ResponseFromModel(&m), nil
}

func (s *Service) FindAppointments(ctx context.Context, hospitalID ulid.ULID) ([]Response, error) {
	var ms []models.Appointment
	if err := s.db.WithContext(ctx).Where("hospital_id = ?", hospitalID).Find(&ms).Error; err != nil {
		return nil, err
	}

	res := make([]Response, len(ms))
	for i, m := range ms {
		res[i] = *ResponseFromModel(&m)
	}
	return res, nil
}

func (s *Service) FindAppointmentByUser(ctx context.Context, userID ulid.ULID, page, limit int) ([]Response, error) {
	var ms []models.Appointment
	offset := (page - 1) * limit
	if err := s.db.WithContext(ctx).
		Where("doctor_id = ? OR patient_id = ?", userID, userID).
		Offset(offset).
		Limit(limit).
		Find(&ms).Error; err != nil {
		return nil, err
	}

	res := make([]Response, len(ms))
	for i, m := range ms {
		res[i] = *ResponseFromModel(&m)
	}
	return res, nil
}
