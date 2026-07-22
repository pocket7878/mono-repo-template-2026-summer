package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrEmptyTodoTitle = errors.New("todo title must not be empty")

type TodoTitle struct {
	value string
}

func NewTodoTitle(value string) (TodoTitle, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return TodoTitle{}, ErrEmptyTodoTitle
	}
	return TodoTitle{value: trimmed}, nil
}

func ReconstructTodoTitle(value string) TodoTitle {
	return TodoTitle{value: value}
}

func (title TodoTitle) String() string {
	return title.value
}

type Todo struct {
	ID        uuid.UUID
	Title     TodoTitle
	Completed bool
	Version   int32
	DueDate   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTodo(
	id uuid.UUID,
	title string,
	completed bool,
	version int32,
	dueDate *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (*Todo, error) {
	todoTitle, err := NewTodoTitle(title)
	if err != nil {
		return nil, err
	}

	return &Todo{
		ID:        id,
		Title:     todoTitle,
		Completed: completed,
		Version:   version,
		DueDate:   dueDate,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func ReconstructTodo(
	id uuid.UUID,
	title string,
	completed bool,
	version int32,
	dueDate *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Todo {
	return &Todo{
		ID:        id,
		Title:     ReconstructTodoTitle(title),
		Completed: completed,
		Version:   version,
		DueDate:   dueDate,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
