package todo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/entity"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/todo"
	"github.com/stretchr/testify/require"
)

type stubRepository struct {
	created   *entity.Todo
	createErr error
	list      []*entity.Todo
	listErr   error
	found     *entity.Todo
	findErr   error
	updated   *entity.Todo
	updateErr error
	deletedID uuid.UUID
	deleteErr error
}

func (repository *stubRepository) Create(_ context.Context, created *entity.Todo) error {
	repository.created = created
	return repository.createErr
}

func (repository *stubRepository) List(context.Context) ([]*entity.Todo, error) {
	return repository.list, repository.listErr
}

func (repository *stubRepository) FindByID(context.Context, uuid.UUID) (*entity.Todo, error) {
	if repository.found != nil || repository.findErr != nil {
		return repository.found, repository.findErr
	}
	return entity.ReconstructTodo(uuid.New(), "Buy milk", false, 3, nil, time.Now(), time.Now()), nil
}

func (repository *stubRepository) Update(_ context.Context, updated *entity.Todo, _ int32) error {
	repository.updated = updated
	return repository.updateErr
}

func (repository *stubRepository) Delete(_ context.Context, id uuid.UUID) error {
	repository.deletedID = id
	return repository.deleteErr
}

func TestCreateBuildsInitialTodo(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	repository := &stubRepository{}
	useCase := todo.NewTodoUseCase(repository, func() time.Time { return now })

	created, err := useCase.Create(context.Background(), todo.CreateInput{Title: "Buy milk"})

	require.NoError(t, err)
	require.Equal(t, repository.created, created)
	require.Equal(t, int32(1), created.Version)
	require.Equal(t, now, created.CreatedAt)
}

func TestCreateReturnsEntityValidationError(t *testing.T) {
	t.Parallel()

	useCase := todo.NewTodoUseCase(&stubRepository{}, time.Now)
	_, err := useCase.Create(context.Background(), todo.CreateInput{Title: " "})

	require.ErrorIs(t, err, entity.ErrEmptyTodoTitle)
}

func TestListReturnsRepositoryTodos(t *testing.T) {
	t.Parallel()

	expected := []*entity.Todo{entity.ReconstructTodo(uuid.New(), "Buy milk", false, 1, nil, time.Now(), time.Now())}
	useCase := todo.NewTodoUseCase(&stubRepository{list: expected}, time.Now)

	actual, err := useCase.List(context.Background())

	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestUpdateAppliesProvidedFields(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 22, 13, 0, 0, 0, time.UTC)
	existing := entity.ReconstructTodo(uuid.New(), "Buy milk", false, 1, nil, now.Add(-time.Hour), now.Add(-time.Hour))
	repository := &stubRepository{found: existing}
	useCase := todo.NewTodoUseCase(repository, func() time.Time { return now })
	title := "Buy oat milk"
	completed := true

	updated, err := useCase.Update(context.Background(), todo.UpdateInput{
		ID:        existing.ID,
		Title:     &title,
		Completed: &completed,
		Version:   1,
	})

	require.NoError(t, err)
	require.Equal(t, "Buy oat milk", updated.Title.String())
	require.True(t, updated.Completed)
	require.Equal(t, int32(2), updated.Version)
	require.Equal(t, updated, repository.updated)
}

func TestDeleteDelegatesToRepository(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	useCase := todo.NewTodoUseCase(repository, time.Now)
	id := uuid.New()

	require.NoError(t, useCase.Delete(context.Background(), id))
	require.Equal(t, id, repository.deletedID)
}

func TestUpdateReturnsRepositoryFindError(t *testing.T) {
	t.Parallel()

	expected := errors.New("database unavailable")
	useCase := todo.NewTodoUseCase(&stubRepository{findErr: expected}, time.Now)

	_, err := useCase.Update(context.Background(), todo.UpdateInput{ID: uuid.New(), Version: 1})

	require.ErrorIs(t, err, expected)
}

func TestDeleteReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	expected := errors.New("database unavailable")
	useCase := todo.NewTodoUseCase(&stubRepository{deleteErr: expected}, time.Now)

	require.ErrorIs(t, useCase.Delete(context.Background(), uuid.New()), expected)
}

func TestUpdateMapsVersionConflict(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{updateErr: todo.ErrVersionConflict}
	useCase := todo.NewTodoUseCase(repository, func() time.Time { return time.Now().UTC() })

	_, err := useCase.Update(context.Background(), todo.UpdateInput{
		ID:      uuid.New(),
		Version: 3,
	})

	require.ErrorIs(t, err, todo.ErrConflict)
	require.NotErrorIs(t, err, errors.New("unrelated error"))
}
