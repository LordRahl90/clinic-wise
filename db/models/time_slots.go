package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

// Timeslot represents a time slot for a user in a hospital
// there might be some adjustments as we don't want to be booking individually. maybe day of the week?
type Timeslot struct {
	ID         ulid.ULID `json:"id" gorm:"size:50"`
	HospitalID ulid.ULID `json:"hospital_id"`
	UserID     ulid.ULID `json:"user_id"`
	Date       string    `json:"date"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`

	gorm.Model
}

func (t *Timeslot) BeforeCreate(_ *gorm.DB) (err error) {
	t.ID = ulid.Make()
	return
}
