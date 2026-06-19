package transcriber

import (
	"context"
	"net"

	"github.com/oklog/ulid/v2"
	"golang.org/x/net/websocket"
)

type noteService interface {
	Update(ctx context.Context, appointmentID ulid.ULID, content string) error
}

type Service struct {
	conn        *websocket.Conn
	noteService noteService
}

func New(client *websocket.Conn, service noteService) *Service {
	return &Service{conn: client, noteService: service}
}

func (s *Service) Start(ctx context.Context, req *StartTranscribingRequest) *net.Conn {
	// we open a streaming connection to the transcriber service.
	// as long as the connection is open, we handle the transcribed response and update the appointment note.
	// when the connection is closed, we return.
	return nil
}

func (s *Service) Complete(ctx context.Context, conn *net.Conn) error {
	// we close the connection to the transcriber service.
	return nil
}
