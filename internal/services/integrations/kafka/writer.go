package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"clinic-wise/internal/entities"

	"github.com/segmentio/kafka-go"
)

type Writer struct {
	writer *kafka.Writer
}

func NewWriter(config *Config) *Writer {

	return &Writer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(config.Brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (s *Writer) Write(ctx context.Context, topic string, event *entities.Event) error {
	slog.InfoContext(ctx, "writing event to queue", "event", event)
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(event.EventID),
		Value: b,
	})
}

func (s *Writer) Close() error {
	return s.writer.Close()
}
