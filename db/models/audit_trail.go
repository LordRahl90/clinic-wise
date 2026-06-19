package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type AuditTrail struct {
	ID            ulid.ULID      `json:"id" gorm:"size:50"`
	ActorID       ulid.ULID      `json:"actor_id" gorm:"size:50"`
	ActorName     string         `json:"actor_name"`
	ActorRole     string         `json:"actor_role"`
	Action        string         `json:"action"`
	EntityType    string         `json:"entity_type"`
	EntityID      string         `json:"entity_id" gorm:"size:50"`
	AppointmentID string         `json:"appointment_id" gorm:"size:50"`
	Message       string         `json:"message" gorm:"type:text"`
	Changes       []byte         `json:"changes" gorm:"type:json"`

	gorm.Model
}

func (a *AuditTrail) BeforeCreate(_ *gorm.DB) (err error) {
	if a.ID == (ulid.ULID{}) {
		a.ID = ulid.Make()
	}
	return
}


