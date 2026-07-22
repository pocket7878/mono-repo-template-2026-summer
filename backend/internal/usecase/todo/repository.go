package todo

import (
	"context"

	"github.com/google/uuid"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/entity"
)

type Repository interface {
	Create(context.Context, *entity.Todo) error
	List(context.Context) ([]*entity.Todo, error)
	FindByID(context.Context, uuid.UUID) (*entity.Todo, error)
	Update(context.Context, *entity.Todo, int32) error
	Delete(context.Context, uuid.UUID) error
}
