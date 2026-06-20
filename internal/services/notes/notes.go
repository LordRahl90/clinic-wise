package notes

import (
	"context"
	"log/slog"
	"time"

	"clinic-wise/db/models"
	"clinic-wise/internal/entities"
	"clinic-wise/internal/services/audittrail"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Writer interface {
	Write(ctx context.Context, event *entities.Event) error
}

type DictationService interface {
	Transcribe(ctx context.Context, audio []byte) (string, error)
}

type Service struct {
	writer           Writer
	dictationService DictationService
	db               *gorm.DB
}

func New(db *gorm.DB, writer Writer) *Service {
	return &Service{
		writer: writer,
		db:     db,
	}
}

func NewNoopWriter() Writer {
	return noopWriter{}
}

func (s *Service) Create(ctx context.Context, req *CreateNoteRequest) (*Response, error) {
	m := req.ToModel()
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}

	if err := s.writer.Write(ctx, makeNoteEvent(m, entities.NoteCreated)); err != nil {
		return nil, err
	}

	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:       m.DoctorID,
		Action:        "note_created",
		EntityType:    "note",
		EntityID:      m.ID.String(),
		AppointmentID: m.AppointmentID.String(),
		Message:       "added note to appointment " + m.AppointmentID.String(),
		Changes: []audittrail.Change{
			{Field: "content", After: m.Content},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record note create audit", "note_id", m.ID.String(), "error", err)
	}

	return FromModel(m), nil
}

func (s *Service) Update(ctx context.Context, userID, noteID ulid.ULID, content string) error {
	var exists models.Note
	if err := s.db.WithContext(ctx).Where("id = ?", noteID).First(&exists).Error; err != nil {
		return err
	}

	// mocked check to make sure that auth is preserved
	if exists.DoctorID != userID {
		return gorm.ErrRecordNotFound
	}
	previousContent := exists.Content
	exists.Content += content

	if err := s.db.WithContext(ctx).Save(&exists).Error; err != nil {
		return err
	}

	if err := audittrail.Record(ctx, s.db, &audittrail.RecordRequest{
		ActorID:       userID,
		Action:        "note_updated",
		EntityType:    "note",
		EntityID:      exists.ID.String(),
		AppointmentID: exists.AppointmentID.String(),
		Message:       "updated note on appointment " + exists.AppointmentID.String(),
		Changes: []audittrail.Change{
			{Field: "content", Before: previousContent, After: exists.Content},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "failed to record note update audit", "note_id", exists.ID.String(), "error", err)
	}

	return s.writer.Write(ctx, makeNoteEvent(&exists, entities.NoteUpdated))
}

func (s *Service) GetAppointmentNotes(ctx context.Context, userID, appointmentID ulid.ULID) ([]Response, error) {
	var notes []models.Note
	if err := s.db.WithContext(ctx).Where("appointment_id = ?", appointmentID).Find(&notes).Error; err != nil {
		return nil, err
	}

	res := make([]Response, len(notes))
	for i, note := range notes {
		res[i] = *FromModel(&note)
	}
	return res, nil
}

func makeNoteEvent(note *models.Note, eventType entities.EventType) *entities.Event {
	payload := entities.NotePayload{
		NoteID:   note.ID.String(),
		NoteText: note.Content,
	}
	evt := &entities.Event{
		EventID:       uuid.NewString(),
		EventType:     eventType,
		Timestamp:     time.Now(),
		AppointmentID: note.AppointmentID,
		PatientID:     note.PatientID,
		Payload:       payload,
	}
	return evt
}

type noopWriter struct{}

func (noopWriter) Write(context.Context, *entities.Event) error {
	return nil
}
