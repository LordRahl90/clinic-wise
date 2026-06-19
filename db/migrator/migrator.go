package migrator

import (
	"clinic-wise/db/models"

	"gorm.io/gorm"
)

var m = []interface{}{
	&models.Hospital{},
	&models.User{},
	&models.Timeslot{},
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(m...); err != nil {
		return err
	}
	return nil
}
