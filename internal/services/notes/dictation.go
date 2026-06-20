package notes

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
)

func (s *Service) StartDictation(ctx context.Context, conn *websocket.Conn, req *StartDictationRequest) error {
	for i := 0; i < 100; i++ {
		_, mt, err := conn.ReadMessage()
		if err != nil {
			// we bail out
			return err
		}
		fmt.Printf("\n\nreceived %s\n\n", mt)

		// transcription service also streams its content, so we have to handle that
		//content, err := s.dictationService.Transcribe(ctx, mt)

		// we read chunks from the connection and pass it on to the transcriber.
		// we gather messages from the transcriber until we get the final message and then append this to the note content.
		if err := conn.WriteJSON([]byte(`{"message":"hello world"}`)); err != nil {
			return err
		}
	}
	return nil
}
