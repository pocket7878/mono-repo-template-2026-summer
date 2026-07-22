package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/entity"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/handler/api"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/todo"
)

type TodoUseCase interface {
	Create(context.Context, todo.CreateInput) (*entity.Todo, error)
	List(context.Context) ([]*entity.Todo, error)
	Update(context.Context, todo.UpdateInput) (*entity.Todo, error)
	Delete(context.Context, uuid.UUID) error
}

type TodoHandler struct {
	useCase TodoUseCase
}

func NewTodoHandler(useCase TodoUseCase) *TodoHandler {
	return &TodoHandler{useCase: useCase}
}

func (handler *TodoHandler) TodosList(c *gin.Context) {
	todos, err := handler.useCase.List(c.Request.Context())
	if err != nil {
		handler.writeProblem(c, http.StatusInternalServerError, "internal-error", "Internal server error", nil)
		return
	}

	response := make([]api.Todo, 0, len(todos))
	for _, item := range todos {
		response = append(response, toAPITodo(item))
	}
	c.JSON(http.StatusOK, response)
}

func (handler *TodoHandler) TodosCreate(c *gin.Context) {
	var body api.CreateTodoRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		handler.writeProblem(c, http.StatusBadRequest, "invalid-request", "Invalid request", nil)
		return
	}

	created, err := handler.useCase.Create(c.Request.Context(), todo.CreateInput{Title: body.Title, DueDate: body.DueDate})
	if err != nil {
		if errors.Is(err, entity.ErrEmptyTodoTitle) {
			handler.writeProblem(c, http.StatusBadRequest, "invalid-todo-title", "Todo title is required", nil)
			return
		}
		handler.writeProblem(c, http.StatusInternalServerError, "internal-error", "Internal server error", nil)
		return
	}

	c.JSON(http.StatusOK, toAPITodo(created))
}

func (handler *TodoHandler) TodosUpdate(c *gin.Context, todoID string) {
	id, err := uuid.Parse(todoID)
	if err != nil {
		handler.writeProblem(c, http.StatusBadRequest, "invalid-todo-id", "Invalid todo ID", nil)
		return
	}

	var body api.UpdateTodoRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		handler.writeProblem(c, http.StatusBadRequest, "invalid-request", "Invalid request", nil)
		return
	}

	updated, err := handler.useCase.Update(c.Request.Context(), todo.UpdateInput{
		ID:        id,
		Title:     body.Title,
		Completed: body.Completed,
		DueDate:   body.DueDate,
		Version:   body.Version,
	})
	if err != nil {
		switch {
		case errors.Is(err, todo.ErrConflict):
			handler.writeProblem(c, http.StatusConflict, "version-conflict", "Todo was modified", nil)
		case errors.Is(err, todo.ErrNotFound):
			handler.writeProblem(c, http.StatusNotFound, "todo-not-found", "Todo not found", nil)
		case errors.Is(err, entity.ErrEmptyTodoTitle):
			handler.writeProblem(c, http.StatusBadRequest, "invalid-todo-title", "Todo title is required", nil)
		default:
			handler.writeProblem(c, http.StatusInternalServerError, "internal-error", "Internal server error", nil)
		}
		return
	}

	c.JSON(http.StatusOK, toAPITodo(updated))
}

func (handler *TodoHandler) TodosDelete(c *gin.Context, todoID string) {
	id, err := uuid.Parse(todoID)
	if err != nil {
		handler.writeProblem(c, http.StatusBadRequest, "invalid-todo-id", "Invalid todo ID", nil)
		return
	}
	if err := handler.useCase.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, todo.ErrNotFound) {
			handler.writeProblem(c, http.StatusNotFound, "todo-not-found", "Todo not found", nil)
			return
		}
		handler.writeProblem(c, http.StatusInternalServerError, "internal-error", "Internal server error", nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (handler *TodoHandler) writeProblem(c *gin.Context, status int, kind string, title string, detail *string) {
	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, api.Problem{
		Type:   "https://reference-todo.dev/problems/" + kind,
		Title:  title,
		Status: int32(status),
		Detail: detail,
	})
}

func toAPITodo(todo *entity.Todo) api.Todo {
	return api.Todo{
		Id:        todo.ID.String(),
		Title:     todo.Title.String(),
		Completed: todo.Completed,
		Version:   todo.Version,
		DueDate:   cloneTime(todo.DueDate),
		CreatedAt: todo.CreatedAt,
		UpdatedAt: todo.UpdatedAt,
	}
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
