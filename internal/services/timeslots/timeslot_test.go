package timeslots

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"clinic-wise/db/migrator"
	"clinic-wise/pkg/testhelper"

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
	svc := New(db)
	req := &CreateTimeslotRequest{
		HospitalID: ulid.Make().String(),
		UserID:     ulid.Make().String(),
		Date:       time.Now().Format(time.DateOnly),
		StartTime:  time.Now().Format(time.TimeOnly),
		EndTime:    time.Now().Add(30 * time.Minute).Format(time.TimeOnly),
	}
	res, err := svc.Create(t.Context(), req)
	require.NoError(t, err)
	require.NotEmpty(t, res.ID)
}

func TestFindByUser(t *testing.T) {
	userID := ulid.Make()
	svc := New(db)

	for i := 0; i < 3; i++ {
		req := &CreateTimeslotRequest{
			UserID:     userID.String(),
			HospitalID: ulid.Make().String(),
			Date:       time.Now().Format(time.DateOnly),
			StartTime:  time.Now().Format(time.TimeOnly),
			EndTime:    time.Now().Add(30 * time.Minute).Format(time.TimeOnly),
		}

		res, err := svc.Create(t.Context(), req)
		require.NoError(t, err)
		require.NotEmpty(t, res.ID)
	}

	result, err := svc.FindByUser(t.Context(), userID)
	require.NoError(t, err)
	require.Equal(t, 3, len(result))
}
