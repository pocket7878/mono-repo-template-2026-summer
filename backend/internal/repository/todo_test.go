package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/repository/internal/db"
	"github.com/stretchr/testify/require"
)

func TestReconstructTodoReconstructsDomainEntity(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	dueDate := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, time.July, 22, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	todo := reconstructTodo(db.Todo{
		ID: pgtype.UUID{Bytes: id, Valid: true}, Title: "Buy milk", Completed: false,
		Version: 3, DueDate: pgtype.Timestamptz{Time: dueDate, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
	})

	require.Equal(t, id, todo.ID)
	require.Equal(t, "Buy milk", todo.Title.String())
	require.Equal(t, int32(3), todo.Version)
	require.Equal(t, dueDate, *todo.DueDate)
}
