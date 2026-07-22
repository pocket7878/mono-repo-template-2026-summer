package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestOutboxRepositoryClaimsAndDeadLettersEvent(t *testing.T) {
	databaseURL := os.Getenv("POSTGRES_TEST_DSN")
	if databaseURL == "" {
		t.Skip("POSTGRES_TEST_DSN is required for PostgreSQL integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	_, err = pool.Exec(ctx, `
DROP TABLE IF EXISTS job_dead_letters;
DROP TABLE IF EXISTS job_outbox_events;
CREATE TABLE job_outbox_events (
  id UUID PRIMARY KEY, event_type TEXT NOT NULL, payload JSONB NOT NULL,
  attempt_count INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'pending',
  last_error TEXT, lease_expires_at TIMESTAMPTZ, published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE job_dead_letters (
  outbox_event_id UUID PRIMARY KEY REFERENCES job_outbox_events(id), event_type TEXT NOT NULL,
  payload JSONB NOT NULL, failure_reason TEXT NOT NULL, failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`)
	require.NoError(t, err)

	id := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO job_outbox_events (id, event_type, payload) VALUES ($1, 'todo_import', '{"csv":"title"}')`, id)
	require.NoError(t, err)

	repository := repository.NewOutboxRepository(pool)
	events, err := repository.ClaimPending(ctx, 10, time.Now().UTC())
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, id, events[0].ID)
	require.Equal(t, 1, events[0].Attempt)
	require.NoError(t, repository.MarkFailed(ctx, id, "broker unavailable"))

	var status string
	var deadLetters int
	require.NoError(t, pool.QueryRow(ctx, `SELECT status FROM job_outbox_events WHERE id = $1`, id).Scan(&status))
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM job_dead_letters WHERE outbox_event_id = $1`, id).Scan(&deadLetters))
	require.Equal(t, "failed", status)
	require.Equal(t, 1, deadLetters)
}

func TestOutboxRepositoryMarksPublishedEvent(t *testing.T) {
	databaseURL := os.Getenv("POSTGRES_TEST_DSN")
	if databaseURL == "" {
		t.Skip("POSTGRES_TEST_DSN is required for PostgreSQL integration tests")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	_, err = pool.Exec(ctx, `
DROP TABLE IF EXISTS job_dead_letters;
DROP TABLE IF EXISTS job_outbox_events;
CREATE TABLE job_outbox_events (
  id UUID PRIMARY KEY, event_type TEXT NOT NULL, payload JSONB NOT NULL,
  attempt_count INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'pending',
  last_error TEXT, lease_expires_at TIMESTAMPTZ, published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE job_dead_letters (
  outbox_event_id UUID PRIMARY KEY REFERENCES job_outbox_events(id), event_type TEXT NOT NULL,
  payload JSONB NOT NULL, failure_reason TEXT NOT NULL, failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`)
	require.NoError(t, err)

	id := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO job_outbox_events (id, event_type, payload) VALUES ($1, 'todo_import', '{}')`, id)
	require.NoError(t, err)
	require.NoError(t, repository.NewOutboxRepository(pool).MarkPublished(ctx, id))

	var status string
	require.NoError(t, pool.QueryRow(ctx, `SELECT status FROM job_outbox_events WHERE id = $1`, id).Scan(&status))
	require.Equal(t, "published", status)
}
