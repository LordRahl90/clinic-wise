package diagnosis

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

func (s *Service) Create(ctx context.Context, req *CreateDiagnosisRequest) (*Response, error) {
	m := req.ToModel()

	var appointment models.Appointment
	if err := s.db.WithContext(ctx).Where("id = ?", m.AppointmentID).First(&appointment).Error; err != nil {
		return nil, err
	}

	if appointment.DoctorID != m.DoctorID || appointment.PatientID != m.PatientID || appointment.HospitalID != m.HospitalID {
		return nil, gorm.ErrRecordNotFound
	}

	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:       m.DoctorID,
		Action:        "diagnosis_created",
		EntityType:    "diagnosis",
		EntityID:      m.ID.String(),
		AppointmentID: m.AppointmentID.String(),
		Message:       "added diagnosis to appointment " + m.AppointmentID.String(),
		Changes: []audittrail.Change{
			{Field: "diagnosis", After: m.Diagnosis},
			{Field: "details", After: m.Details},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record diagnosis create audit", "diagnosis_id", m.ID.String(), "error", err)
	}

	return FromModel(m), nil
}

func (s *Service) Find(ctx context.Context, userID, diagnosisID ulid.ULID) (*Response, error) {
	var m models.Diagnosis
	if err := s.db.WithContext(ctx).Where("id = ?", diagnosisID).First(&m).Error; err != nil {
		return nil, err
	}

	if m.DoctorID != userID && m.PatientID != userID {
		return nil, gorm.ErrRecordNotFound
	}

	return FromModel(&m), nil
}

func (s *Service) Dismiss(ctx context.Context, doctorID, diagnosisID ulid.ULID) (*Response, error) {
	var m models.Diagnosis
	if err := s.db.WithContext(ctx).Where("id = ?", diagnosisID).First(&m).Error; err != nil {
		return nil, err
	}

	if m.DoctorID != doctorID {
		return nil, gorm.ErrRecordNotFound
	}

	m.Dismissed = true
	if err := s.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, err
	}
	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:       doctorID,
		Action:        "diagnosis_dismissed",
		EntityType:    "diagnosis",
		EntityID:      m.ID.String(),
		AppointmentID: m.AppointmentID.String(),
		Message:       "dismissed diagnosis for appointment " + m.AppointmentID.String(),
		Changes: []audittrail.Change{
			{Field: "dismissed", Before: false, After: true},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record diagnosis dismiss audit", "diagnosis_id", m.ID.String(), "error", err)
	}

	return FromModel(&m), nil
}

func (s *Service) FindByUser(ctx context.Context, userID ulid.ULID) ([]Response, error) {
	var ms []models.Diagnosis
	if err := s.db.WithContext(ctx).
		Where("doctor_id = ? OR patient_id = ?", userID, userID).
		Find(&ms).Error; err != nil {
		return nil, err
	}

	res := make([]Response, len(ms))
	for i, m := range ms {
		res[i] = *FromModel(&m)
	}
	return res, nil
}


