package audittrail

import (
	"clinic-wise/db/models"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"
)

type Response struct {
	ID            ulid.ULID        `json:"id"`
	ActorID       ulid.ULID        `json:"actor_id"`
	ActorName     string           `json:"actor_name"`
	ActorRole     string           `json:"actor_role"`
	Action        string           `json:"action"`
	EntityType    string           `json:"entity_type"`
	EntityID      string           `json:"entity_id"`
	AppointmentID string           `json:"appointment_id"`
	Message       string           `json:"message"`
	Changes       json.RawMessage  `json:"changes"`
	CreatedAt     time.Time        `json:"created_at"`
}

func FromModel(m *models.AuditTrail) Response {
	return Response{
		ID:            m.ID,
		ActorID:       m.ActorID,
		ActorName:     m.ActorName,
		ActorRole:     m.ActorRole,
		Action:        m.Action,
		EntityType:    m.EntityType,
		EntityID:      m.EntityID,
		AppointmentID: m.AppointmentID,
		Message:       m.Message,
		Changes:       json.RawMessage(m.Changes),
		CreatedAt:     m.CreatedAt,
	}
}

