package kafka

import (
	"context"
	"log/slog"

	"clinic-wise/internal/entities"
)

type Config struct{}

type Service struct {
	config *Config
}

func NewService(config *Config) *Service {
	return &Service{
		config: config,
	}
}

func (s *Service) Write(ctx context.Context, event *entities.Event) error {
	slog.InfoContext(ctx, "writing event to queue", "event", event)

	return nil
}

func (s *Service) Read(ctx context.Context) chan *entities.Event {
	slog.InfoContext(ctx, "reading event from queue")
	return nil
}
