package prescriptions

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"clinic-wise/db/migrator"
	"clinic-wise/db/models"
	"clinic-wise/internal/entities"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockEventWriter struct {
	writeFunc func(ctx context.Context, event *entities.Event) error
}

func (m *mockEventWriter) Write(ctx context.Context, event *entities.Event) error {
	return m.writeFunc(ctx, event)
}

var (
	db *gorm.DB
)

func TestMain(m *testing.M) {
	code := 1
	container, err := testhelper.GetMySQLContainer(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := container.Terminate(context.TODO()); err != nil {
			log.Fatal(err)
		}
		os.Exit(code)
	}()

	db = testhelper.SetupContainerTestDB(context.TODO(), container)
	if err := migrator.Migrate(db); err != nil {
		log.Fatal(err)
	}

	code = m.Run()
}

func TestService_CreateFindAndFindByAppointment(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	outsider := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, outsider)

	hospitalID := ulid.Make()
	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  hospitalID,
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "checkup",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	writerCallCount := 0
	svc := New(db, &mockEventWriter{
		writeFunc: func(ctx context.Context, event *entities.Event) error {
			writerCallCount++
			require.Equal(t, entities.PrescriptionCreated, event.EventType)
			require.Equal(t, appointment.ID, event.AppointmentID)
			require.Equal(t, patient.ID, event.PatientID)
			payload, ok := event.Payload.(entities.PrescriptionPayload)
			require.True(t, ok)
			require.Equal(t, "Take one tablet daily", payload.Medication)
			return nil
		},
	})

	created, err := svc.Create(t.Context(), &CreatePrescriptionRequest{
		HospitalID:     hospitalID.String(),
		DoctorID:       doctor.ID.String(),
		PatientID:      patient.ID.String(),
		AppointmentID:  appointment.ID.String(),
		ExpirationDate: time.Now().Add(48 * time.Hour),
		Details:        "Take one tablet daily",
	})
	require.NoError(t, err)
	require.Equal(t, models.ActivePrescription, created.Status)
	require.Equal(t, 1, writerCallCount)

	// Doctor and patient can read prescription details.
	doctorRead, err := svc.Find(t.Context(), doctor.ID, ulid.MustParse(created.ID))
	require.NoError(t, err)
	require.Equal(t, created.ID, doctorRead.ID)

	patientRead, err := svc.Find(t.Context(), patient.ID, ulid.MustParse(created.ID))
	require.NoError(t, err)
	require.Equal(t, created.ID, patientRead.ID)

	// Unrelated users cannot read prescription details.
	_, err = svc.Find(t.Context(), outsider.ID, ulid.MustParse(created.ID))
	require.Error(t, err)

	doctorList, err := svc.FindByAppointment(t.Context(), doctor.ID, appointment.ID)
	require.NoError(t, err)
	require.Len(t, doctorList, 1)

	patientList, err := svc.FindByAppointment(t.Context(), patient.ID, appointment.ID)
	require.NoError(t, err)
	require.Len(t, patientList, 1)

	_, err = svc.FindByAppointment(t.Context(), outsider.ID, appointment.ID)
	require.Error(t, err)
}

func TestService_Dispatch(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	pharmacist := testhelper.CreateUser(db, models.Pharmacist)
	require.NotNil(t, pharmacist)

	hospitalID := ulid.Make()
	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  hospitalID,
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "checkup",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	writerCallCount := 0
	svc := New(db, &mockEventWriter{
		writeFunc: func(ctx context.Context, event *entities.Event) error {
			writerCallCount++
			switch writerCallCount {
			case 1:
				require.Equal(t, entities.PrescriptionCreated, event.EventType)
			case 2:
				require.Equal(t, entities.PrescriptionUpdated, event.EventType)
				payload, ok := event.Payload.(entities.PrescriptionPayload)
				require.True(t, ok)
				require.Equal(t, "Use twice daily", payload.Medication)
			default:
				t.Fatalf("unexpected event call count %d", writerCallCount)
			}
			return nil
		},
	})

	created, err := svc.Create(t.Context(), &CreatePrescriptionRequest{
		HospitalID:     hospitalID.String(),
		DoctorID:       doctor.ID.String(),
		PatientID:      patient.ID.String(),
		AppointmentID:  appointment.ID.String(),
		ExpirationDate: time.Now().Add(24 * time.Hour),
		Details:        "Use twice daily",
	})
	require.NoError(t, err)

	_, err = svc.Dispatch(t.Context(), doctor.ID, ulid.MustParse(created.ID))
	require.Error(t, err)

	dispatched, err := svc.Dispatch(t.Context(), pharmacist.ID, ulid.MustParse(created.ID))
	require.NoError(t, err)
	require.Equal(t, models.Unavailable, dispatched.Status)
	require.Equal(t, 2, writerCallCount)

	storedPrescriptionID := ulid.MustParse(created.ID)
	storedBytes, err := storedPrescriptionID.MarshalBinary()
	require.NoError(t, err)

	var stored models.Prescription
	require.NoError(t, db.Where("id = ?", storedBytes).First(&stored).Error)
	require.Equal(t, models.Unavailable, stored.Status)

	_, err = svc.Dispatch(t.Context(), pharmacist.ID, ulid.MustParse(created.ID))
	require.ErrorIs(t, err, ErrPrescriptionUnavailable)
}

func TestService_DispatchRejectsExpiredPrescription(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	pharmacist := testhelper.CreateUser(db, models.Pharmacist)
	require.NotNil(t, pharmacist)

	hospitalID := ulid.Make()
	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  hospitalID,
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "expired checkup",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	writerCallCount := 0
	svc := New(db, &mockEventWriter{
		writeFunc: func(ctx context.Context, event *entities.Event) error {
			writerCallCount++
			require.Equal(t, entities.PrescriptionCreated, event.EventType)
			return nil
		},
	})

	created, err := svc.Create(t.Context(), &CreatePrescriptionRequest{
		HospitalID:     hospitalID.String(),
		DoctorID:       doctor.ID.String(),
		PatientID:      patient.ID.String(),
		AppointmentID:  appointment.ID.String(),
		ExpirationDate: time.Now().Add(-1 * time.Hour),
		Details:        "Expired prescription",
	})
	require.NoError(t, err)
	require.Equal(t, 1, writerCallCount)

	_, err = svc.Dispatch(t.Context(), pharmacist.ID, ulid.MustParse(created.ID))
	require.ErrorIs(t, err, ErrPrescriptionExpired)
	require.Equal(t, 1, writerCallCount)

	storedPrescriptionID := ulid.MustParse(created.ID)
	storedBytes, err := storedPrescriptionID.MarshalBinary()
	require.NoError(t, err)

	var stored models.Prescription
	require.NoError(t, db.Where("id = ?", storedBytes).First(&stored).Error)
	require.Equal(t, models.ActivePrescription, stored.Status)
}

func TestService_CreateContinuesWhenEventWriteFails(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)

	hospitalID := ulid.Make()
	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  hospitalID,
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "event failure create",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	svc := New(db, &mockEventWriter{
		writeFunc: func(ctx context.Context, event *entities.Event) error {
			return errors.New("queue unavailable")
		},
	})

	created, err := svc.Create(t.Context(), &CreatePrescriptionRequest{
		HospitalID:     hospitalID.String(),
		DoctorID:       doctor.ID.String(),
		PatientID:      patient.ID.String(),
		AppointmentID:  appointment.ID.String(),
		ExpirationDate: time.Now().Add(12 * time.Hour),
		Details:        "Create should still succeed",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	storedPrescriptionID := ulid.MustParse(created.ID)
	storedBytes, err := storedPrescriptionID.MarshalBinary()
	require.NoError(t, err)

	var stored models.Prescription
	require.NoError(t, db.Where("id = ?", storedBytes).First(&stored).Error)
	require.Equal(t, models.ActivePrescription, stored.Status)
}

func TestService_DispatchContinuesWhenEventWriteFails(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	pharmacist := testhelper.CreateUser(db, models.Pharmacist)
	require.NotNil(t, pharmacist)

	hospitalID := ulid.Make()
	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  hospitalID,
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "event failure dispatch",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	writerCalls := 0
	svc := New(db, &mockEventWriter{
		writeFunc: func(ctx context.Context, event *entities.Event) error {
			writerCalls++
			if writerCalls == 1 {
				return nil
			}
			return errors.New("queue unavailable")
		},
	})

	created, err := svc.Create(t.Context(), &CreatePrescriptionRequest{
		HospitalID:     hospitalID.String(),
		DoctorID:       doctor.ID.String(),
		PatientID:      patient.ID.String(),
		AppointmentID:  appointment.ID.String(),
		ExpirationDate: time.Now().Add(12 * time.Hour),
		Details:        "Dispatch should still succeed",
	})
	require.NoError(t, err)

	dispatched, err := svc.Dispatch(t.Context(), pharmacist.ID, ulid.MustParse(created.ID))
	require.NoError(t, err)
	require.Equal(t, models.Unavailable, dispatched.Status)

	storedPrescriptionID := ulid.MustParse(created.ID)
	storedBytes, err := storedPrescriptionID.MarshalBinary()
	require.NoError(t, err)

	var stored models.Prescription
	require.NoError(t, db.Where("id = ?", storedBytes).First(&stored).Error)
	require.Equal(t, models.Unavailable, stored.Status)
}
