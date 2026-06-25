package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"clinic-wise/internal/entities"

	"github.com/segmentio/kafka-go"
)

type Reader struct {
	reader *kafka.Reader
	topic  string
}

func NewReader(config *Config) *Reader {
	// we might need to isolate the readerConfig if config.Auth is true
	// different implementation uses different auth technique eg SCRAM
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		MaxBytes: 10e3,
		Topic:    config.Topic,
		GroupID:  config.ConsumerGroup,
	})
	return &Reader{
		reader: reader,
		topic:  config.Topic,
	}
}

func (r *Reader) Close() error {
	return r.reader.Close()
}

// Reads from the queue and passes it on to a channel
func (r *Reader) Read(ctx context.Context, topic string, ch chan *entities.Event) error {
	slog.InfoContext(ctx, "reading event from queue", "topic", topic)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := r.reader.FetchMessage(ctx)
			if err != nil {
				return err
			}
			var event entities.Event
			err = json.Unmarshal(msg.Value, &event)
			if err != nil {
				slog.ErrorContext(ctx, "failed to unmarshal event", "error", err)
				continue
			}
			ch <- &event
			// we can commit to committing messages later on
		}
	}
}
