package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestNewTodoRejectsBlankTitle(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	_, err := entity.NewTodo(uuid.New(), "   ", false, 1, nil, now, now)

	require.ErrorIs(t, err, entity.ErrEmptyTodoTitle)
}

func TestNewTodoTrimsTitle(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	todo, err := entity.NewTodo(uuid.New(), "  Buy milk  ", false, 1, nil, now, now)

	require.NoError(t, err)
	require.Equal(t, "Buy milk", todo.Title.String())
}

func TestReconstructTodoPreservesPersistedValues(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	dueDate := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	todo := entity.ReconstructTodo(id, "Stored title", true, 4, &dueDate, createdAt, createdAt.Add(time.Minute))

	require.Equal(t, id, todo.ID)
	require.Equal(t, "Stored title", todo.Title.String())
	require.True(t, todo.Completed)
	require.Equal(t, int32(4), todo.Version)
	require.Equal(t, dueDate, *todo.DueDate)
}
