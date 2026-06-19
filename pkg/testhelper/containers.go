package testhelper

import (
	"context"

	mysqlModule "github.com/testcontainers/testcontainers-go/modules/mysql"
)

// GetMySQLContainer returns mysql container
func GetMySQLContainer(ctx context.Context) (*mysqlModule.MySQLContainer, error) {
	return mysqlModule.Run(ctx, "mysql:8.0",
		mysqlModule.WithDatabase("metis"),
		mysqlModule.WithUsername("root"),
		mysqlModule.WithPassword("rootpassword"),
	)
}
