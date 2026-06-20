package notes

import (
	"clinic-wise/db/migrator"
	"clinic-wise/db/models"
	"clinic-wise/internal/entities"
	"clinic-wise/pkg/testhelper"
	"context"
	"log"
	"log/slog"
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
	if err := migrator.MigrateUp(db); err != nil {
		log.Fatal(err)
	}

	code = m.Run()
}

func TestService_Create(t *testing.T) {
	doctor := testhelper.CreateUser(db, models.Doctor)
	require.NotNil(t, doctor)
	patient := testhelper.CreateUser(db, models.Patient)
	require.NotNil(t, patient)

	req := &CreateNoteRequest{
		HospitalID:    ulid.Make(),
		PatientID:     patient.ID,
		DoctorID:      doctor.ID,
		AppointmentID: ulid.Make(),
		Content:       "Test note content",
	}

	mockedWriter := &mockWriter{
		writeFunc: func(ctx context.Context, event *entities.Event) error {
			slog.InfoContext(ctx, "writing stuff", "event", event)
			require.NotEmpty(t, event.EventID)
			require.Equal(t, entities.NoteCreated, event.EventType)
			require.Equal(t, req.PatientID, event.PatientID)
			require.Equal(t, req.AppointmentID, event.AppointmentID)
			require.NotNil(t, event.Payload)
			//require.Equal(t, req.Content, event.Payload.NoteText)
			return nil
		},
	}
	svc := New(db, mockedWriter)

	res, err := svc.Create(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
}

type mockWriter struct {
	writeFunc func(ctx context.Context, event *entities.Event) error
}

func (m *mockWriter) Write(ctx context.Context, event *entities.Event) error {
	return m.writeFunc(ctx, event)
}
