package webhooks

import (
	"bytes"
	"clinic-wise/db/models"
	"clinic-wise/internal/entities"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	client *http.Client
}

func New(db *gorm.DB, client *http.Client) *Service {
	return &Service{
		db:     db,
		client: client,
	}
}

func (s *Service) Register(ctx context.Context, req *RegisterWebhookRequest) (*Response, error) {
	model := req.ToModel()
	if err := s.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	return FromModel(model), nil
}

// Trigger
// Ideally, this should be read from a queue and pushed out to the registered webhooks endpoint.
// however, for the sake of this test, we might use channels
func Trigger(ctx context.Context, client *http.Client, db *gorm.DB, eventChan chan *entities.Event) error {
	slog.InfoContext(ctx, "triggering webhooks")
	for msg := range eventChan {
		err := dispatch(ctx, client, db, *msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func dispatch(ctx context.Context, client *http.Client, db *gorm.DB, event entities.Event) error {
	// get the webhook entry for the patient
	var webhook models.Webhook
	if err := db.WithContext(ctx).Where("patient_id = ?", event.PatientID).First(&webhook).Error; err != nil {
		return err
	}

	code, err := sendRequest(ctx, client, webhook.URL, event)
	if code != http.StatusOK {
		return err
	}
	return nil
}

// we return the status code and error. the status code is returned so we can implement retry and backoff logic
func sendRequest(ctx context.Context, client *http.Client, path string, payload interface{}) (int, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	// we don't need to read the body
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return resp.StatusCode, nil
}
