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
	HospitalID ulid.ULID `json:"hospital_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Password   string    `json:"password"`
	Role       UserRole  `json:"role"`

	gorm.Model
}

//func (u User) BeforeCreate(_ *gorm.DB) (err error) {
//	u.ID = ulid.Make()
//	return
//}
