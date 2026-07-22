package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/entity"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/todo"
)

type TodoRepository struct {
	pool *pgxpool.Pool
}

func NewTodoRepository(pool *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{pool: pool}
}

func (repository *TodoRepository) Create(ctx context.Context, item *entity.Todo) error {
	_, err := repository.pool.Exec(ctx, `
INSERT INTO todos (id, title, completed, version, due_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		item.ID,
		item.Title.String(),
		item.Completed,
		item.Version,
		item.DueDate,
		item.CreatedAt,
		item.UpdatedAt,
	)
	return err
}

func (repository *TodoRepository) List(ctx context.Context) ([]*entity.Todo, error) {
	rows, err := repository.pool.Query(ctx, `
SELECT id, title, completed, version, due_date, created_at, updated_at
FROM todos
ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := make([]*entity.Todo, 0)
	for rows.Next() {
		item, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
}

func (repository *TodoRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Todo, error) {
	item, err := scanTodo(repository.pool.QueryRow(ctx, `
SELECT id, title, completed, version, due_date, created_at, updated_at
FROM todos
WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, todo.ErrNotFound
	}
	return item, err
}

func (repository *TodoRepository) Update(ctx context.Context, item *entity.Todo, expectedVersion int32) error {
	result, err := repository.pool.Exec(ctx, `
UPDATE todos
SET title = $1, completed = $2, version = $3, due_date = $4, updated_at = $5
WHERE id = $6 AND version = $7`,
		item.Title.String(),
		item.Completed,
		item.Version,
		item.DueDate,
		item.UpdatedAt,
		item.ID,
		expectedVersion,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return todo.ErrVersionConflict
	}
	return nil
}

func (repository *TodoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := repository.pool.Exec(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return todo.ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanTodo(row rowScanner) (*entity.Todo, error) {
	var (
		id        uuid.UUID
		title     string
		completed bool
		version   int32
		dueDate   *time.Time
		createdAt time.Time
		updatedAt time.Time
	)
	if err := row.Scan(&id, &title, &completed, &version, &dueDate, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	return entity.ReconstructTodo(id, title, completed, version, dueDate, createdAt, updatedAt), nil
}
