package prescriptions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"clinic-wise/db/models"
	"clinic-wise/internal/entities"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type EventWriter interface {
	Write(ctx context.Context, event *entities.Event) error
}

type Service struct {
	db     *gorm.DB
	writer EventWriter
}

var (
	ErrPrescriptionExpired     = errors.New("prescription has expired")
	ErrPrescriptionUnavailable = errors.New("prescription is unavailable")
)

func New(db *gorm.DB, writer EventWriter) *Service {
	return &Service{db: db, writer: writer}
}

func NewNoopEventWriter() EventWriter {
	return noopEventWriter{}
}

func (s *Service) Create(ctx context.Context, req *CreatePrescriptionRequest) (*Response, error) {
	m, err := req.ToModel()
	if err != nil {
		return nil, err
	}

	appointmentID, err := ulid.ParseStrict(req.AppointmentID)
	if err != nil {
		return nil, err
	}
	appointmentBytes, err := appointmentID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var appointment models.Appointment
	if err := s.db.WithContext(ctx).Where("id = ?", appointmentBytes).First(&appointment).Error; err != nil {
		return nil, err
	}

	if appointment.DoctorID != m.DoctorID || appointment.PatientID != m.PatientID || appointment.HospitalID != m.HospitalID {
		return nil, gorm.ErrRecordNotFound
	}

	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	if err := s.writer.Write(ctx, makePrescriptionEvent(m, entities.PrescriptionCreated)); err != nil {
		slog.ErrorContext(ctx, "failed to emit prescription event", "event_type", entities.PrescriptionCreated, "prescription_id", m.ID.String(), "error", err)
	}

	return FromModel(m), nil
}

func (s *Service) Dispatch(ctx context.Context, pharmacistID, prescriptionID ulid.ULID) (*Response, error) {
	pharmacistBytes, err := pharmacistID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var pharmacist models.User
	if err := s.db.WithContext(ctx).Where("id = ?", pharmacistBytes).First(&pharmacist).Error; err != nil {
		return nil, err
	}
	if pharmacist.Role != models.Pharmacist {
		return nil, fmt.Errorf("only pharmacists can dispatch prescriptions")
	}

	var prescription models.Prescription
	if err := s.db.WithContext(ctx).Where("id = ?", prescriptionID).First(&prescription).Error; err != nil {
		return nil, err
	}

	if !prescription.ExpirationDate.After(time.Now()) {
		return nil, ErrPrescriptionExpired
	}
	if prescription.Status == models.Unavailable {
		return nil, ErrPrescriptionUnavailable
	}

	prescription.Status = models.Unavailable
	if err := s.db.WithContext(ctx).Save(&prescription).Error; err != nil {
		return nil, err
	}
	if err := s.writer.Write(ctx, makePrescriptionEvent(&prescription, entities.PrescriptionUpdated)); err != nil {
		slog.ErrorContext(ctx, "failed to emit prescription event", "event_type", entities.PrescriptionUpdated, "prescription_id", prescription.ID.String(), "error", err)
	}

	return FromModel(&prescription), nil
}

func (s *Service) Find(ctx context.Context, userID, prescriptionID ulid.ULID) (*Response, error) {
	prescriptionBytes, err := prescriptionID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var m models.Prescription
	if err := s.db.WithContext(ctx).Where("id = ?", prescriptionBytes).First(&m).Error; err != nil {
		return nil, err
	}

	if m.DoctorID != userID && m.PatientID != userID {
		return nil, gorm.ErrRecordNotFound
	}

	return FromModel(&m), nil
}

func (s *Service) FindByAppointment(ctx context.Context, userID, appointmentID ulid.ULID) ([]Response, error) {
	appointmentBytes, err := appointmentID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var appointment models.Appointment
	if err := s.db.WithContext(ctx).Where("id = ?", appointmentBytes).First(&appointment).Error; err != nil {
		return nil, err
	}

	if appointment.DoctorID != userID && appointment.PatientID != userID {
		return nil, gorm.ErrRecordNotFound
	}

	var ms []models.Prescription
	if err := s.db.WithContext(ctx).Where("appointment_id = ?", appointment.ID).Find(&ms).Error; err != nil {
		return nil, err
	}

	res := make([]Response, len(ms))
	for i, m := range ms {
		res[i] = *FromModel(&m)
	}
	return res, nil
}

type noopEventWriter struct{}

func (noopEventWriter) Write(context.Context, *entities.Event) error {
	return nil
}

func makePrescriptionEvent(prescription *models.Prescription, eventType entities.EventType) *entities.Event {
	payload := entities.PrescriptionPayload{
		PrescriptionID: prescription.ID.String(),
		Medication:     prescription.Details,
		ExpiresAt:      prescription.ExpirationDate,
	}

	return &entities.Event{
		EventID:       uuid.NewString(),
		EventType:     eventType,
		Timestamp:     time.Now(),
		AppointmentID: prescription.AppointmentID,
		PatientID:     prescription.PatientID,
		Payload:       payload,
	}
}
