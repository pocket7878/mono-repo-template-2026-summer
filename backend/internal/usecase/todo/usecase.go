package todo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/entity"
)

type Clock func() time.Time

type TodoUseCase struct {
	repository Repository
	clock      Clock
}

func NewTodoUseCase(repository Repository, clock Clock) *TodoUseCase {
	return &TodoUseCase{repository: repository, clock: clock}
}

type CreateInput struct {
	Title   string
	DueDate *time.Time
}

func (useCase *TodoUseCase) Create(ctx context.Context, input CreateInput) (*entity.Todo, error) {
	now := useCase.clock()
	created, err := entity.NewTodo(uuid.New(), input.Title, false, 1, input.DueDate, now, now)
	if err != nil {
		return nil, err
	}
	if err := useCase.repository.Create(ctx, created); err != nil {
		return nil, err
	}
	return created, nil
}

func (useCase *TodoUseCase) List(ctx context.Context) ([]*entity.Todo, error) {
	return useCase.repository.List(ctx)
}

type UpdateInput struct {
	ID        uuid.UUID
	Title     *string
	Completed *bool
	DueDate   *time.Time
	Version   int32
}

func (useCase *TodoUseCase) Update(ctx context.Context, input UpdateInput) (*entity.Todo, error) {
	existing, err := useCase.repository.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	title := existing.Title.String()
	if input.Title != nil {
		title = *input.Title
	}
	completed := existing.Completed
	if input.Completed != nil {
		completed = *input.Completed
	}
	dueDate := existing.DueDate
	if input.DueDate != nil {
		dueDate = input.DueDate
	}

	updated, err := entity.NewTodo(
		existing.ID,
		title,
		completed,
		existing.Version+1,
		dueDate,
		existing.CreatedAt,
		useCase.clock(),
	)
	if err != nil {
		return nil, err
	}
	if err := useCase.repository.Update(ctx, updated, input.Version); err != nil {
		if errors.Is(err, ErrVersionConflict) {
			return nil, ErrConflict
		}
		return nil, err
	}
	return updated, nil
}

func (useCase *TodoUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return useCase.repository.Delete(ctx, id)
}
