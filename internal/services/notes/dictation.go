package notes

import (
	"clinic-wise/db/models"
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
)

type Message struct {
	Type    string `json:"type"`
	Content []byte `json:"content"`
}

func (s *Service) StartDictation(ctx context.Context, conn *websocket.Conn) error {
	fmt.Printf("\n\nStarting here\n\n")
	for i := 0; i < 100; i++ {
		_, mt, err := conn.ReadMessage()
		if err != nil {
			// we bail out
			return err
		}
		slog.InfoContext(ctx, "received message", "type", mt)

		// transcription service also streams its content, so we have to handle that
		//content, err := s.dictationService.Transcribe(ctx, mt)

		// we read chunks from the connection and pass it on to the transcriber.
		// we gather messages from the transcriber until we get the final message and then append this to the note content.
		// if err := conn.WriteJSON([]byte(`{"message":"hello world"}`)); err != nil {
		if err := conn.WriteJSON(string(mt)); err != nil {
			return err
		}
	}
	return nil
}

var (
	feedBackMap = make(map[ulid.ULID][]NoteFeeback)
)

// Feedback this handles dictation feedback from the STT provider
func (s *Service) Feedback(ctx context.Context, feedback NoteFeeback) error {
	feedBackMap[feedback.AppointmentID] = append(feedBackMap[feedback.AppointmentID], feedback)
	if !feedback.IsFinal {
		return nil
	}

	// this means the feedback is final, so we collect it and sort it by sequence, then we update the note content for the appointment.
	v, ok := feedBackMap[feedback.AppointmentID]
	if !ok {
		return fmt.Errorf("no feedback found for appointment %s", feedback.AppointmentID.String())
	}

	// sort by sequence
	sort.Slice(v, func(i, j int) bool {
		return v[i].Sequence < v[j].Sequence
	})

	var output strings.Builder
	for _, f := range v {
		output.WriteString(f.Text)
		output.WriteString(" ")
	}

	var appt *models.Appointment
	if err := s.db.WithContext(ctx).Where("id = ?", feedback.AppointmentID).First(&appt).Error; err != nil {
		return err
	}

	return s.UpdateWithAppointmentID(ctx, appt.DoctorID, feedback.AppointmentID, output.String())
}
