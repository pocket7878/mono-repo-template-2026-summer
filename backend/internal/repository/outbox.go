package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/repository/internal/db"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/outbox"
)

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

func (repository *OutboxRepository) ClaimPending(ctx context.Context, limit int, now time.Time) ([]outbox.Event, error) {
	rows, err := db.New(repository.pool).ClaimPendingOutboxEvents(ctx, db.ClaimPendingOutboxEventsParams{
		LeaseExpiresAt: pgTimeValue(now.Add(time.Minute)), Now: pgTimeValue(now), BatchLimit: int32(limit),
	})
	if err != nil {
		return nil, err
	}
	events := make([]outbox.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, outbox.Event{ID: uuid.UUID(row.ID.Bytes), Type: row.EventType, Payload: row.Payload, Attempt: int(row.AttemptCount)})
	}
	return events, nil
}

func (repository *OutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	return db.New(repository.pool).MarkOutboxPublished(ctx, pgUUID(id))
}

func (repository *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := db.New(tx)
	event, err := queries.LockOutboxEvent(ctx, pgUUID(id))
	if err != nil {
		return err
	}
	if err := queries.MarkOutboxFailed(ctx, db.MarkOutboxFailedParams{ID: pgUUID(id), LastError: pgtype.Text{String: reason, Valid: true}}); err != nil {
		return err
	}
	if err := queries.InsertDeadLetter(ctx, db.InsertDeadLetterParams{
		OutboxEventID: pgUUID(id), EventType: event.EventType, Payload: event.Payload, FailureReason: reason,
	}); err != nil {
		return fmt.Errorf("insert dead letter: %w", err)
	}
	return tx.Commit(ctx)
}

func pgTimeValue(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}
