package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type UserRole string

const (
	Admin      UserRole = "admin"
	Patient    UserRole = "patient"
	Doctor     UserRole = "doctor"
	Pharmacist UserRole = "pharmacist"
)

type User struct {
	ID         ulid.ULID `json:"id" gorm:"size:50"`
	HospitalID ulid.ULID `json:"hospital_id" gorm:"size:50"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	Password   string    `json:"-" gorm:"type:varchar(255)"`
	Role       UserRole  `json:"role"`
	Accepted   bool      `json:"accepted" gorm:"default:false"`

	gorm.Model
}

func (u *User) BeforeCreate(_ *gorm.DB) (err error) {
	if u.ID == (ulid.ULID{}) {
		u.ID = ulid.Make()
	}
	return
}
