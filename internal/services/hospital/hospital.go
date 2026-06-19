package hospital

import (
	"context"

	"clinic-wise/db/models"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, req *CreateHospitalRequest) (*CreateHospitalResponse, error) {
	hospital := req.ToModel()
	if err := s.db.WithContext(ctx).Create(hospital).Error; err != nil {
		return nil, err
	}

	return &CreateHospitalResponse{
		ID: hospital.ID.String(),
	}, nil
}

func (s *Service) Stats(ctx context.Context, hospitalID ulid.ULID) (*StatsResponse, error) {
	var totalAppointments int64
	if err := s.db.WithContext(ctx).
		Model(&models.Appointment{}).
		Where("hospital_id = ?", hospitalID).
		Count(&totalAppointments).Error; err != nil {
		return nil, err
	}

	var activePatients int64
	if err := s.db.WithContext(ctx).
		Model(&models.Appointment{}).
		Distinct("patient_id").
		Where("hospital_id = ? AND status IN ?", hospitalID, []models.AppointmentStatus{
			models.AppointmentStatusActive, models.AppointmentStatusConfirmed,
		}).
		Count(&activePatients).Error; err != nil {
		return nil, err
	}

	var prescriptionCount int64
	if err := s.db.WithContext(ctx).
		Model(&models.Prescription{}).
		Where("hospital_id = ?", hospitalID).
		Count(&prescriptionCount).Error; err != nil {
		return nil, err
	}

	return &StatsResponse{
		TotalAppointments: totalAppointments,
		ActivePatients:    activePatients,
		PrescriptionCount: prescriptionCount,
	}, nil
}
