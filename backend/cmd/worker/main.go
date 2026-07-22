package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/outbox"
)

type logPublisher struct{}

func (logPublisher) Publish(_ context.Context, event outbox.Event) error {
	log.Printf("published outbox event id=%s type=%s attempt=%d", event.ID, event.Type, event.Attempt)
	return nil
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	worker := outbox.NewWorker(newOutboxRepository(pool), logPublisher{}, 3)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := worker.RunOnce(context.Background()); err != nil {
			log.Printf("outbox worker run failed: %v", err)
		}
	}
}
