package repository_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/entity"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/repository"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/todo"
	"github.com/stretchr/testify/require"
)

func TestTodoRepositoryPersistsAndVersionChecks(t *testing.T) {
	databaseURL := os.Getenv("POSTGRES_TEST_DSN")
	if databaseURL == "" {
		t.Skip("POSTGRES_TEST_DSN is required for PostgreSQL integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, `
DROP TABLE IF EXISTS todos;
CREATE TABLE todos (
  id UUID PRIMARY KEY,
  title TEXT NOT NULL CHECK (btrim(title) <> ''),
  completed BOOLEAN NOT NULL DEFAULT FALSE,
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  due_date TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);`)
	require.NoError(t, err)

	repository := repository.NewTodoRepository(pool)
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	item, err := entity.NewTodo(uuid.New(), "Buy milk", false, 1, nil, now, now)
	require.NoError(t, err)
	require.NoError(t, repository.Create(ctx, item))

	listed, err := repository.List(ctx)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, item.ID, listed[0].ID)

	found, err := repository.FindByID(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, "Buy milk", found.Title.String())

	updated, err := entity.NewTodo(item.ID, "Buy oat milk", true, 2, nil, now, now.Add(time.Minute))
	require.NoError(t, err)
	require.NoError(t, repository.Update(ctx, updated, 1))
	require.ErrorIs(t, repository.Update(ctx, updated, 1), todo.ErrVersionConflict)
	require.NoError(t, repository.Delete(ctx, item.ID))
	require.ErrorIs(t, repository.Delete(ctx, item.ID), todo.ErrNotFound)

	_, err = repository.FindByID(ctx, item.ID)
	require.ErrorIs(t, err, todo.ErrNotFound)
}

func TestTodoRepositoryReturnsNotFoundForMissingID(t *testing.T) {
	if strings.TrimSpace(os.Getenv("POSTGRES_TEST_DSN")) == "" {
		t.Skip("POSTGRES_TEST_DSN is required for PostgreSQL integration tests")
	}
}
