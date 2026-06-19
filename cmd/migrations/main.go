package main

import (
	"clinic-wise/db"
	"clinic-wise/db/migrator"
	"clinic-wise/db/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/oklog/ulid/v2"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	config := db.Config{
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
	}

	dbase, err := db.New(&config)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	if err := migrator.Migrate(dbase); err != nil {
		log.Fatal("Failed to migrate", err)
	}

	us := models.User{
		ID: ulid.Make(),
	}

	if err := dbase.Create(&us).Error; err != nil {
		log.Fatal("Failed to create user", err)
	}

	log.Println("Migration completed successfully")
}
