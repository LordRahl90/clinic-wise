package hospital

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"clinic-wise/db/migrator"
	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

var (
	db *sql.DB
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

	db = testhelper.SetupContainerTestDBForSQL(context.TODO(), container)
	if err := migrator.MigrateUp(db); err != nil {
		log.Fatal(err)
	}

	code = m.Run()
}

func TestCreate(t *testing.T) {
	svc := New(db)
	req := &CreateHospitalRequest{
		Name: "Test Hospital",
	}
	res, err := svc.Create(t.Context(), req)
	require.NoError(t, err)
	require.NotEmpty(t, res.ID)

	id, err := ulid.Parse(res.ID)
	require.NoError(t, err)
	require.Equal(t, req.Name, res.Name)

	result, err := svc.Find(t.Context(), id)
	require.NoError(t, err)
	require.Equal(t, res.ID, result.ID)
	require.Equal(t, res.Name, result.Name)
}
