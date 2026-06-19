package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Hospital struct {
	ID   ulid.ULID `json:"id" gorm:"size:50"`
	Name string    `json:"name"`
	gorm.Model

	Users []User `json:"users"`
}

func (h Hospital) BeforeCreate(_ *gorm.DB) (err error) {
	h.ID = ulid.Make()
	return
}
