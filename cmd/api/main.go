package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"clinic-wise/db/migrator"
	"clinic-wise/internal/server"
	"clinic-wise/internal/services/integrations/queue"
	"clinic-wise/internal/services/webhooks"

	"clinic-wise/db"

	"github.com/joho/godotenv"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dbConfig := &db.Config{
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
	}

	database, err := db.New(dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	// we might want this migration to happen independently though.
	if err := migrator.Migrate(database); err != nil {
		log.Fatal(err)
	}

	writer := queue.New()

	serverConfig := &server.Config{
		DB:            database,
		Port:          os.Getenv("PORT"),
		SigningSecret: os.Getenv("SIGNING_SECRET"),
		Writer:        writer,
	}
	apiServer := server.New(serverConfig)

	errChan := make(chan error)
	go func() {
		errChan <- apiServer.Run()
	}()

	go func() {
		// ideally, this would be a separate service consuming from a queue. For the sake of this test, we will read from the channel directly.
		client := &http.Client{
			Timeout: 10 * time.Second,
		} // placeholder. I will eventually use a client that allows retry and backoff logic
		errChan <- webhooks.Trigger(ctx, client, database, writer.Read(ctx))
	}()

	select {
	case <-ctx.Done():
		slog.InfoContext(ctx, "Shutting down server...")
	case err := <-errChan:
		if err != nil {
			slog.ErrorContext(ctx, "server failed with error", "error", err)
		}
	}

	writer.Close(ctx)
}
