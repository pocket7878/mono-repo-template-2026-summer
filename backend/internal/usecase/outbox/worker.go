package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID      uuid.UUID
	Type    string
	Payload []byte
	Attempt int
}

type Repository interface {
	ClaimPending(context.Context, int, time.Time) ([]Event, error)
	MarkPublished(context.Context, uuid.UUID) error
	MarkFailed(context.Context, uuid.UUID, string) error
}

type Publisher interface {
	Publish(context.Context, Event) error
}

type Worker struct {
	repository  Repository
	publisher   Publisher
	maxAttempts int
}

func NewWorker(repository Repository, publisher Publisher, maxAttempts int) *Worker {
	return &Worker{repository: repository, publisher: publisher, maxAttempts: maxAttempts}
}

func (worker *Worker) RunOnce(ctx context.Context) error {
	events, err := worker.repository.ClaimPending(ctx, 100, time.Now().UTC())
	if err != nil {
		return err
	}
	for _, event := range events {
		if err := worker.publisher.Publish(ctx, event); err == nil {
			if err := worker.repository.MarkPublished(ctx, event.ID); err != nil {
				return err
			}
		} else if event.Attempt >= worker.maxAttempts {
			if err := worker.repository.MarkFailed(ctx, event.ID, err.Error()); err != nil {
				return err
			}
		}
	}
	return nil
}
