package hospital

import (
	"clinic-wise/db/migrator"
	"clinic-wise/pkg/testhelper"
	"context"
	"log"
	"os"
	"testing"

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

func TestCreate(t *testing.T) {
	svc := New(db)
	req := &CreateHospitalRequest{
		Name: "Test Hospital",
	}
	err := svc.Create(t.Context(), req)
	require.NoError(t, err)
}
