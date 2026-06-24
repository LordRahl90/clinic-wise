package audittrail

import (
	"clinic-wise/db/migrator"
	"clinic-wise/db/models"
	"clinic-wise/pkg/testhelper"
	"context"
	"log"
	"os"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)
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

func TestRecord(t *testing.T) {
	actor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, actor)

	appointmentID := ulid.Make()
	err := Record(t.Context(), db, &RecordRequest{
		ActorID:       actor.ID,
		Action:        "note_created",
		EntityType:    "note",
		EntityID:      ulid.Make().String(),
		AppointmentID: appointmentID.String(),
		Message:       "added note to appointment " + appointmentID.String(),
		Changes: []Change{
			{Field: "content", After: "Patient is stable"},
		},
	})
	require.NoError(t, err)

	var saved models.AuditTrail
	require.NoError(t, db.Where("appointment_id = ?", appointmentID.String()).First(&saved).Error)
	require.Equal(t, actor.ID, saved.ActorID)
	require.Equal(t, "note_created", saved.Action)
	require.Contains(t, saved.Message, actor.FirstName)
	require.Contains(t, saved.Message, "added note to appointment")
	require.NotEmpty(t, saved.Changes)
}

func TestFindByAppointment(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	otherDoctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, otherDoctor)

	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  ulid.Make(),
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "audit trail appointment",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	require.NoError(t, db.Create(&models.AuditTrail{
		ID:            ulid.Make(),
		ActorID:       doctor.ID,
		ActorName:     "Doctor",
		ActorRole:     string(models.Doctor),
		Action:        "appointment_created",
		EntityType:    "appointment",
		EntityID:      appointment.ID.String(),
		AppointmentID: appointment.ID.String(),
		Message:       "Doctor created appointment " + appointment.ID.String(),
		Changes:       []byte("[]"),
	}).Error)

	svc := New(db)
	logs, err := svc.FindByAppointment(t.Context(), doctor.ID, appointment.ID, FilterQuery{})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, appointment.ID.String(), logs[0].AppointmentID)

	patientLogs, err := svc.FindByAppointment(t.Context(), patient.ID, appointment.ID, FilterQuery{})
	require.NoError(t, err)
	require.Len(t, patientLogs, 1)

	_, err = svc.FindByAppointment(t.Context(), otherDoctor.ID, appointment.ID, FilterQuery{})
	require.ErrorIs(t, err, ErrForbidden)
}

func TestFindByEntity(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)
	unrelated := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, unrelated)

	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  ulid.Make(),
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "entity audit test",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	noteID := ulid.Make().String()
	require.NoError(t, db.Create(&models.AuditTrail{
		ID:            ulid.Make(),
		ActorID:       doctor.ID,
		ActorName:     "Doctor",
		ActorRole:     string(models.Doctor),
		Action:        "note_created",
		EntityType:    "note",
		EntityID:      noteID,
		AppointmentID: appointment.ID.String(),
		Message:       "Doctor added note to appointment",
		Changes:       []byte("[]"),
	}).Error)

	svc := New(db)

	// Doctor (linked via appointment) can view
	logs, err := svc.FindByEntity(t.Context(), doctor.ID, "note", noteID, FilterQuery{})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, noteID, logs[0].EntityID)

	// Patient (linked via appointment) can view
	patientLogs, err := svc.FindByEntity(t.Context(), patient.ID, "note", noteID, FilterQuery{})
	require.NoError(t, err)
	require.Len(t, patientLogs, 1)

	// Unrelated user cannot view
	_, err = svc.FindByEntity(t.Context(), unrelated.ID, "note", noteID, FilterQuery{})
	require.ErrorIs(t, err, ErrForbidden)
}

func TestFindByAppointmentFilters(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)

	appointment := &models.Appointment{
		ID:          ulid.Make(),
		HospitalID:  ulid.Make(),
		DoctorID:    doctor.ID,
		PatientID:   patient.ID,
		TimeslotID:  ulid.Make(),
		Description: "filter test",
		Status:      models.AppointmentStatusActive,
	}
	require.NoError(t, db.Create(appointment).Error)

	createEntry := func(action, entityType string, actorID ulid.ULID) {
		require.NoError(t, db.Create(&models.AuditTrail{
			ID:            ulid.Make(),
			ActorID:       actorID,
			ActorName:     "Actor",
			ActorRole:     "doctor",
			Action:        action,
			EntityType:    entityType,
			EntityID:      ulid.Make().String(),
			AppointmentID: appointment.ID.String(),
			Message:       action + " on appointment",
			Changes:       []byte("[]"),
		}).Error)
	}

	createEntry("note_created", "note", doctor.ID)
	createEntry("note_updated", "note", doctor.ID)
	createEntry("prescription_created", "prescription", doctor.ID)

	svc := New(db)

	// Filter by action
	noteCreated, err := svc.FindByAppointment(t.Context(), doctor.ID, appointment.ID, FilterQuery{Action: "note_created"})
	require.NoError(t, err)
	require.Len(t, noteCreated, 1)
	require.Equal(t, "note_created", noteCreated[0].Action)

	// Filter by actor_id
	actorFiltered, err := svc.FindByAppointment(t.Context(), doctor.ID, appointment.ID, FilterQuery{ActorID: doctor.ID})
	require.NoError(t, err)
	require.Len(t, actorFiltered, 3)

	// Pagination: page=1, limit=2
	page1, err := svc.FindByAppointment(t.Context(), doctor.ID, appointment.ID, FilterQuery{Page: 1, Limit: 2})
	require.NoError(t, err)
	require.Len(t, page1, 2)

	// Pagination: page=2, limit=2
	page2, err := svc.FindByAppointment(t.Context(), doctor.ID, appointment.ID, FilterQuery{Page: 2, Limit: 2})
	require.NoError(t, err)
	require.Len(t, page2, 1)
}

