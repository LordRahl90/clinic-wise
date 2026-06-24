package appointments

import (
	"context"
	"log/slog"

	"clinic-wise/db/models"
	"clinic-wise/internal/services/audittrail"

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

	actorID, err := ulid.ParseStrict(req.ActorID)
	if err == nil {
		err = audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
			ActorID:       actorID,
			Action:        "appointment_created",
			EntityType:    "appointment",
			EntityID:      m.ID.String(),
			AppointmentID: m.ID.String(),
			Message:       "created appointment " + m.ID.String(),
			Changes: []audittrail.Change{
				{Field: "status", After: m.Status},
				{Field: "doctor_id", After: m.DoctorID.String()},
				{Field: "patient_id", After: m.PatientID.String()},
			},
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to record appointment create audit", "appointment_id", m.ID.String(), "error", err)
		}
	}

	return ResponseFromModel(m), nil
}

// Complete marks an appointment as completed by the assigned doctor.
func (s *Service) Complete(ctx context.Context, userID, id ulid.ULID) (*Response, error) {
	var m models.Appointment
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}

	if m.DoctorID != userID {
		return nil, gorm.ErrRecordNotFound
	}

	previousStatus := m.Status
	m.Status = models.AppointmentStatusCompleted
	if err := s.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, err
	}

	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:       userID,
		Action:        "appointment_completed",
		EntityType:    "appointment",
		EntityID:      m.ID.String(),
		AppointmentID: m.ID.String(),
		Message:       "completed appointment " + m.ID.String(),
		Changes: []audittrail.Change{
			{Field: "status", Before: previousStatus, After: m.Status},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record appointment complete audit", "appointment_id", m.ID.String(), "error", err)
	}

	return ResponseFromModel(&m), nil
}

// Find finds an appointment by ID
// this can only be retrieved by the patient or doctor so a userID is required
func (s *Service) Find(ctx context.Context, userID, id ulid.ULID) (*Response, error) {
	var m models.Appointment
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
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
