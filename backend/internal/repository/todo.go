package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/entity"
	db "github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/repository/internal/db"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/todo"
)

type TodoRepository struct {
	pool *pgxpool.Pool
}

func NewTodoRepository(pool *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{pool: pool}
}

func (repository *TodoRepository) Create(ctx context.Context, item *entity.Todo) error {
	return repository.withTransaction(ctx, func(tx pgx.Tx) error {
		queries := db.New(tx)
		if err := queries.CreateTodo(ctx, db.CreateTodoParams{
			ID: pgUUID(item.ID), Title: item.Title.String(), Completed: item.Completed,
			Version: item.Version, DueDate: pgTime(item.DueDate), CreatedAt: pgTime(&item.CreatedAt), UpdatedAt: pgTime(&item.UpdatedAt),
		}); err != nil {
			return err
		}
		return insertTodoEvent(ctx, queries, "todo.created", todoEventPayload(item))
	})
}

func (repository *TodoRepository) List(ctx context.Context) ([]*entity.Todo, error) {
	rows, err := db.New(repository.pool).ListTodos(ctx)
	if err != nil {
		return nil, err
	}
	todos := make([]*entity.Todo, 0, len(rows))
	for _, row := range rows {
		todos = append(todos, reconstructTodo(row))
	}
	return todos, nil
}

func (repository *TodoRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Todo, error) {
	row, err := db.New(repository.pool).FindTodoByID(ctx, pgUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, todo.ErrNotFound
	}
	return reconstructTodo(row), err
}

func (repository *TodoRepository) Update(ctx context.Context, item *entity.Todo, expectedVersion int32) error {
	return repository.withTransaction(ctx, func(tx pgx.Tx) error {
		queries := db.New(tx)
		rows, err := queries.UpdateTodo(ctx, db.UpdateTodoParams{
			Title: item.Title.String(), Completed: item.Completed, Version: item.Version,
			DueDate: pgTime(item.DueDate), UpdatedAt: pgTime(&item.UpdatedAt), ID: pgUUID(item.ID), Version_2: expectedVersion,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return todo.ErrVersionConflict
		}
		return insertTodoEvent(ctx, queries, "todo.updated", todoEventPayload(item))
	})
}

func (repository *TodoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return repository.withTransaction(ctx, func(tx pgx.Tx) error {
		queries := db.New(tx)
		rows, err := queries.DeleteTodo(ctx, pgUUID(id))
		if err != nil {
			return err
		}
		if rows == 0 {
			return todo.ErrNotFound
		}
		return insertTodoEvent(ctx, queries, "todo.deleted", struct {
			ID uuid.UUID `json:"id"`
		}{ID: id})
	})
}

func (repository *TodoRepository) withTransaction(ctx context.Context, operation func(pgx.Tx) error) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := operation(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func insertTodoEvent(ctx context.Context, queries *db.Queries, eventType string, payload any) error {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s outbox payload: %w", eventType, err)
	}
	return queries.InsertOutboxEvent(ctx, db.InsertOutboxEventParams{ID: pgUUID(uuid.New()), EventType: eventType, Payload: encodedPayload})
}

type todoPayload struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Completed bool       `json:"completed"`
	Version   int32      `json:"version"`
	DueDate   *time.Time `json:"dueDate,omitempty"`
}

func todoEventPayload(item *entity.Todo) todoPayload {
	return todoPayload{ID: item.ID, Title: item.Title.String(), Completed: item.Completed, Version: item.Version, DueDate: item.DueDate}
}

func pgUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

func pgTime(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}

func reconstructTodo(row db.Todo) *entity.Todo {
	var dueDate *time.Time
	if row.DueDate.Valid {
		dueDate = &row.DueDate.Time
	}
	return entity.ReconstructTodo(uuid.UUID(row.ID.Bytes), row.Title, row.Completed, row.Version, dueDate, row.CreatedAt.Time, row.UpdatedAt.Time)
}
