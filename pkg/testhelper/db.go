package testhelper

import (
	"context"
	"fmt"
	"log"

	mysqlModule "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// SetupContainerTestDB initializes a test db using the provided container
func SetupContainerTestDB(ctx context.Context, container *mysqlModule.MySQLContainer) *gorm.DB {
	dsn, err := container.ConnectionString(ctx)
	if err != nil {
		log.Fatal(err)
	}
	db, err := gorm.Open(mysql.Open(fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", dsn)), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	return db
}
