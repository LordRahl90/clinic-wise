package audittrail

import (
	"context"
	"errors"
	"time"

	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

// ErrForbidden is returned when the caller is not authorised to view the requested records.
var ErrForbidden = errors.New("access forbidden")

// FilterQuery carries optional filters and pagination for audit trail queries.
type FilterQuery struct {
	Action  string
	ActorID ulid.ULID
	From    time.Time
	To      time.Time
	Page    int // 1-based
	Limit   int // default 50, max 200
}

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) FindByAppointment(ctx context.Context, userID, appointmentID ulid.ULID, filter FilterQuery) ([]Response, error) {
	var appointment models.Appointment
	if err := s.db.WithContext(ctx).Where("id = ?", appointmentID).First(&appointment).Error; err != nil {
		return nil, err
	}
	if appointment.DoctorID != userID && appointment.PatientID != userID {
		return nil, ErrForbidden
	}

	q := s.db.WithContext(ctx).Where("appointment_id = ?", appointmentID.String())
	q = applyFilters(q, filter)

	var logs []models.AuditTrail
	if err := q.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return toResponses(logs), nil
}

func (s *Service) FindByEntity(ctx context.Context, userID ulid.ULID, entityType, entityID string, filter FilterQuery) ([]Response, error) {
	// Load audit records for the entity first so we can determine ownership.
	var sample []models.AuditTrail
	if err := s.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Limit(1).
		Find(&sample).Error; err != nil {
		return nil, err
	}

	if !s.callerMayViewEntity(ctx, userID, entityType, entityID, sample) {
		return nil, ErrForbidden
	}

	q := s.db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", entityType, entityID)
	q = applyFilters(q, filter)

	var logs []models.AuditTrail
	if err := q.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return toResponses(logs), nil
}

// callerMayViewEntity returns true when the caller is related to the entity or was the actor.
func (s *Service) callerMayViewEntity(ctx context.Context, userID ulid.ULID, entityType, entityID string, sample []models.AuditTrail) bool {
	// If we found audit records that have an appointment_id, verify via appointment ownership.
	if len(sample) > 0 && sample[0].AppointmentID != "" {
		apptID, err := ulid.ParseStrict(sample[0].AppointmentID)
		if err == nil {
			var appt models.Appointment
			if err := s.db.WithContext(ctx).Where("id = ?", apptID).First(&appt).Error; err == nil {
				if appt.DoctorID == userID || appt.PatientID == userID {
					return true
				}
			}
		}
	}

	// Allow actors to view their own records for entities without an appointment link.
	if len(sample) > 0 && sample[0].ActorID == userID {
		return true
	}

	// Direct ownership by entity type for entities that don't carry an appointment ID.
	switch entityType {
	case "user":
		var u models.User
		if err := s.db.WithContext(ctx).Where("id = ?", entityID).First(&u).Error; err == nil {
			return u.ID == userID
		}
	case "hospital":
		// Only admins view hospital-level audits; let caller access their own actions.
		if len(sample) > 0 {
			for _, s := range sample {
				if s.ActorID == userID {
					return true
				}
			}
		}
	}
	return false
}

func applyFilters(q *gorm.DB, f FilterQuery) *gorm.DB {
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.ActorID != (ulid.ULID{}) {
		q = q.Where("actor_id = ?", f.ActorID)
	}
	if !f.From.IsZero() {
		q = q.Where("created_at >= ?", f.From)
	}
	if !f.To.IsZero() {
		q = q.Where("created_at <= ?", f.To)
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return q.Limit(limit).Offset(offset)
}

func toResponses(logs []models.AuditTrail) []Response {
	result := make([]Response, len(logs))
	for i, item := range logs {
		result[i] = FromModel(&item)
	}
	return result
}
