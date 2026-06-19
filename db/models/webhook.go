package models

import "github.com/oklog/ulid/v2"

type Webhook struct {
	ID         ulid.ULID `json:"id"`
	HospitalID ulid.ULID `json:"hospital_id"`
	PatientID  ulid.ULID `json:"patient_id"`
	URL        string    `json:"url"`
}
