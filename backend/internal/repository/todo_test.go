package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type stubRow struct {
	values []any
	err    error
}

func (row stubRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return errors.New("unexpected number of scan destinations")
	}
	for index, destination := range destinations {
		switch pointer := destination.(type) {
		case *uuid.UUID:
			*pointer = row.values[index].(uuid.UUID)
		case *string:
			*pointer = row.values[index].(string)
		case *bool:
			*pointer = row.values[index].(bool)
		case *int32:
			*pointer = row.values[index].(int32)
		case **time.Time:
			*pointer = row.values[index].(*time.Time)
		case *time.Time:
			*pointer = row.values[index].(time.Time)
		default:
			return errors.New("unexpected scan destination")
		}
	}
	return nil
}

func TestScanTodoReconstructsDomainEntity(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	dueDate := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, time.July, 22, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	todo, err := scanTodo(stubRow{values: []any{
		id, "Buy milk", false, int32(3), &dueDate, createdAt, updatedAt,
	}})

	require.NoError(t, err)
	require.Equal(t, id, todo.ID)
	require.Equal(t, "Buy milk", todo.Title.String())
	require.Equal(t, int32(3), todo.Version)
	require.Equal(t, dueDate, *todo.DueDate)
}
