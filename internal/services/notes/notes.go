package notes

import (
	"context"
	"time"

	"clinic-wise/db/models"
	"clinic-wise/internal/entities"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Writer interface {
	Write(ctx context.Context, event *entities.Event) error
}

type Service struct {
	writer Writer
	db     *gorm.DB
}

func New(db *gorm.DB, writer Writer) *Service {
	return &Service{
		writer: writer,
		db:     db,
	}
}

func (s *Service) Create(ctx context.Context, req *CreateNoteRequest) (*Response, error) {
	m := req.ToModel()
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}

	if err := s.writer.Write(ctx, makeNoteEvent(m, entities.NoteCreated)); err != nil {
		return nil, err
	}

	return FromModel(m), nil
}

func (s *Service) Update(ctx context.Context, userID, noteID ulid.ULID, content string) error {
	var exists models.Note
	if err := s.db.WithContext(ctx).First(&exists, noteID).Error; err != nil {
		return err
	}

	// mocked check to make sure that auth is preserved
	if exists.DoctorID != userID {
		return gorm.ErrRecordNotFound
	}
	exists.Content += content

	if err := s.db.WithContext(ctx).Save(&exists).Error; err != nil {
		return err
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
