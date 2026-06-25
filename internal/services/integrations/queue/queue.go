package queue

import (
	"context"
	"log/slog"

	"clinic-wise/internal/entities"
)

// Writer this mocks an interface where we write to a kafka/rabbit queue
//type Writer interface {
//	Write(ctx context.Context, event *entities.Event)
//}

type Service struct {
	store chan *entities.Event
}

func New() *Service {
	return &Service{
		store: make(chan *entities.Event),
	}
}

func (s *Service) Write(ctx context.Context, topic string, event *entities.Event) error {
	slog.InfoContext(ctx, "writing event to queue", "topic", topic, "event", event)
	s.store <- event
	return nil
}

func (s *Service) Read(ctx context.Context) chan *entities.Event {
	slog.InfoContext(ctx, "reading event from queue")
	return s.store
}

func (s *Service) Close(ctx context.Context) {
	slog.InfoContext(ctx, "closing store")
	close(s.store)
}
