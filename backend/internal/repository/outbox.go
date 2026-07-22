package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/outbox"
)

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

func (repository *OutboxRepository) ClaimPending(ctx context.Context, limit int, now time.Time) ([]outbox.Event, error) {
	rows, err := repository.pool.Query(ctx, `
UPDATE job_outbox_events
SET attempt_count = attempt_count + 1, lease_expires_at = $1, updated_at = $2
WHERE id IN (
  SELECT id FROM job_outbox_events
  WHERE status = 'pending' AND (lease_expires_at IS NULL OR lease_expires_at < $2)
  ORDER BY created_at, id
  FOR UPDATE SKIP LOCKED
  LIMIT $3
)
RETURNING id, event_type, payload, attempt_count`, now.Add(time.Minute), now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]outbox.Event, 0)
	for rows.Next() {
		var event outbox.Event
		if err := rows.Scan(&event.ID, &event.Type, &event.Payload, &event.Attempt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (repository *OutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := repository.pool.Exec(ctx, `UPDATE job_outbox_events SET status = 'published', published_at = NOW(), updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (repository *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var eventType string
	var payload []byte
	if err := tx.QueryRow(ctx, `SELECT event_type, payload FROM job_outbox_events WHERE id = $1 FOR UPDATE`, id).Scan(&eventType, &payload); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE job_outbox_events SET status = 'failed', last_error = $2, updated_at = NOW() WHERE id = $1`, id, reason); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO job_dead_letters (outbox_event_id, event_type, payload, failure_reason) VALUES ($1, $2, $3, $4)`, id, eventType, payload, reason); err != nil {
		return fmt.Errorf("insert dead letter: %w", err)
	}
	return tx.Commit(ctx)
}
