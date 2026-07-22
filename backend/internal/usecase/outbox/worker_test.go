package outbox_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/outbox"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	events    []outbox.Event
	claimErr  error
	published []uuid.UUID
	failed    []uuid.UUID
}

func (repository *fakeRepository) ClaimPending(context.Context, int, time.Time) ([]outbox.Event, error) {
	return repository.events, repository.claimErr
}

func (repository *fakeRepository) MarkPublished(_ context.Context, id uuid.UUID) error {
	repository.published = append(repository.published, id)
	return nil
}

func (repository *fakeRepository) MarkFailed(_ context.Context, id uuid.UUID, _ string) error {
	repository.failed = append(repository.failed, id)
	return nil
}

type fakePublisher struct{ err error }

func (publisher fakePublisher) Publish(context.Context, outbox.Event) error { return publisher.err }

func TestWorkerMarksPublishedEvent(t *testing.T) {
	t.Parallel()
	event := outbox.Event{ID: uuid.New(), Type: "todo_import", Attempt: 1}
	repository := &fakeRepository{events: []outbox.Event{event}}
	worker := outbox.NewWorker(repository, fakePublisher{}, 3)

	require.NoError(t, worker.RunOnce(context.Background()))
	require.Equal(t, []uuid.UUID{event.ID}, repository.published)
	require.Empty(t, repository.failed)
}

func TestWorkerMovesEventToDeadLetterAfterFinalAttempt(t *testing.T) {
	t.Parallel()
	event := outbox.Event{ID: uuid.New(), Type: "todo_import", Attempt: 3}
	repository := &fakeRepository{events: []outbox.Event{event}}
	worker := outbox.NewWorker(repository, fakePublisher{err: errors.New("broker unavailable")}, 3)

	require.NoError(t, worker.RunOnce(context.Background()))
	require.Empty(t, repository.published)
	require.Equal(t, []uuid.UUID{event.ID}, repository.failed)
}

func TestWorkerReturnsClaimFailure(t *testing.T) {
	t.Parallel()

	expected := errors.New("database unavailable")
	worker := outbox.NewWorker(&fakeRepository{claimErr: expected}, fakePublisher{}, 3)

	require.ErrorIs(t, worker.RunOnce(context.Background()), expected)
}
