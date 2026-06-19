package audittrail

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Change struct {
	Field  string      `json:"field"`
	Before interface{} `json:"before,omitempty"`
	After  interface{} `json:"after,omitempty"`
}

type RecordRequest struct {
	ActorID       ulid.ULID
	Action        string
	EntityType    string
	EntityID      string
	AppointmentID string
	Message       string
	Changes       []Change
}

func Record(ctx context.Context, db *gorm.DB, req *RecordRequest) error {
	if db == nil || req == nil {
		return nil
	}

	actorName, actorRole := resolveActor(ctx, db, req.ActorID)
	changesJSON, err := json.Marshal(req.Changes)
	if err != nil {
		changesJSON = []byte("[]")
	}

	entry := &models.AuditTrail{
		ActorID:       req.ActorID,
		ActorName:     actorName,
		ActorRole:     actorRole,
		Action:        req.Action,
		EntityType:    req.EntityType,
		EntityID:      req.EntityID,
		AppointmentID: req.AppointmentID,
		Message:       toMessage(actorName, req),
		Changes:       changesJSON,
	}

	return db.WithContext(ctx).Create(entry).Error
}

func resolveActor(ctx context.Context, db *gorm.DB, actorID ulid.ULID) (string, string) {
	if actorID == (ulid.ULID{}) {
		return "System", "system"
	}

	var actor models.User
	if err := db.WithContext(ctx).Where("id = ?", actorID).First(&actor).Error; err != nil {
		return "Unknown User", "unknown"
	}

	fullName := strings.TrimSpace(actor.FirstName + " " + actor.LastName)
	if fullName == "" {
		fullName = actor.Email
	}
	if fullName == "" {
		fullName = actor.ID.String()
	}
	return fullName, string(actor.Role)
}

func toMessage(actorName string, req *RecordRequest) string {
	if req.Message != "" {
		return fmt.Sprintf("%s %s", actorName, req.Message)
	}

	message := "performed an action"
	if req.Action != "" {
		message = strings.ReplaceAll(req.Action, "_", " ")
	}
	if req.EntityType != "" && req.EntityID != "" {
		return fmt.Sprintf("%s %s on %s %s", actorName, message, req.EntityType, req.EntityID)
	}
	return fmt.Sprintf("%s %s", actorName, message)
}


