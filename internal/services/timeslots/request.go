package timeslots

import (
	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
)

type CreateTimeslotRequest struct {
	HospitalID string `json:"hospital_id"`
	UserID     string `json:"user_id"`
	Date       string `json:"date"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
}

func (c *CreateTimeslotRequest) ToModel() (*models.Timeslot, error) {
	hospitalID, err := ulid.Parse(c.HospitalID)
	if err != nil {
		return nil, err
	}

	userID, err := ulid.Parse(c.UserID)
	if err != nil {
		return nil, err
	}

	return &models.Timeslot{
		HospitalID: hospitalID,
		UserID:     userID,
		Date:       c.Date,
		StartTime:  c.StartTime,
		EndTime:    c.EndTime,
	}, nil
}
