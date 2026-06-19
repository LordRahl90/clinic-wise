package notes

import (
	"context"

	"github.com/oklog/ulid/v2"
)

func (s *Service) StartDictation(ctx context.Context, userID, appointmentID ulid.ULID) error {
	// we upgrade the connection to a websocket connection and start receiving the transcribed response from the transcriber service.
	// we store the transcribed response in the database.

	return nil
}
