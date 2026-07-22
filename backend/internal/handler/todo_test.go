package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/entity"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/handler"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/handler/api"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/todo"
	"github.com/stretchr/testify/require"
)

type conflictUseCase struct{}

func (conflictUseCase) Create(context.Context, todo.CreateInput) (*entity.Todo, error) {
	return nil, nil
}
func (conflictUseCase) List(context.Context) ([]*entity.Todo, error) { return nil, nil }
func (conflictUseCase) Update(context.Context, todo.UpdateInput) (*entity.Todo, error) {
	return nil, todo.ErrConflict
}
func (conflictUseCase) Delete(context.Context, uuid.UUID) error { return nil }

type stubUseCase struct {
	createResult *entity.Todo
	createErr    error
	listResult   []*entity.Todo
	listErr      error
	updateResult *entity.Todo
	updateErr    error
	deleteErr    error
}

func (stub stubUseCase) Create(context.Context, todo.CreateInput) (*entity.Todo, error) {
	return stub.createResult, stub.createErr
}
func (stub stubUseCase) List(context.Context) ([]*entity.Todo, error) {
	return stub.listResult, stub.listErr
}
func (stub stubUseCase) Update(context.Context, todo.UpdateInput) (*entity.Todo, error) {
	return stub.updateResult, stub.updateErr
}
func (stub stubUseCase) Delete(context.Context, uuid.UUID) error { return stub.deleteErr }

func TestTodosUpdateReturnsProblemDetailsForVersionConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(conflictUseCase{}))

	request := httptest.NewRequest(
		http.MethodPatch,
		"/todos/41bb7dd8-20f5-4bd7-93af-2b7163bb7a35",
		strings.NewReader(`{"version":3}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusConflict, response.Code)
	require.Equal(t, "application/problem+json", response.Header().Get("Content-Type"))

	var problem api.Problem
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &problem))
	require.Equal(t, int32(http.StatusConflict), problem.Status)
	require.Equal(t, "https://reference-todo.dev/problems/version-conflict", problem.Type)
	require.False(t, errors.Is(errors.New(problem.Title), todo.ErrConflict))
}

func TestTodosCreateReturnsCreatedTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	created := entity.ReconstructTodo(uuid.New(), "Buy milk", false, 1, nil, now, now)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{createResult: created}))
	request := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title":"Buy milk"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"id":"`+created.ID.String()+`","title":"Buy milk","completed":false,"version":1,"createdAt":"2026-07-22T12:00:00Z","updatedAt":"2026-07-22T12:00:00Z"}`, response.Body.String())
}

func TestTodosCreateRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{}))
	request := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestTodosCreateMapsTitleValidationToBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{createErr: entity.ErrEmptyTodoTitle}))
	request := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title":" "}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "invalid-todo-title")
}

func TestTodosListReturnsTodos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	item := entity.ReconstructTodo(uuid.New(), "Buy milk", false, 1, nil, now, now)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{listResult: []*entity.Todo{item}}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/todos", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "Buy milk")
}

func TestTodosListMapsFailureToInternalProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{listErr: errors.New("database unavailable")}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/todos", nil))

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), "internal-error")
}

func TestTodosUpdateReturnsUpdatedTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	updated := entity.ReconstructTodo(uuid.New(), "Buy milk", true, 2, nil, now, now)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{updateResult: updated}))
	request := httptest.NewRequest(http.MethodPatch, "/todos/41bb7dd8-20f5-4bd7-93af-2b7163bb7a35", strings.NewReader(`{"completed":true,"version":1}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "Buy milk")
}

func TestTodosUpdateRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{}))
	request := httptest.NewRequest(http.MethodPatch, "/todos/not-a-uuid", strings.NewReader(`{"version":1}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestTodosUpdateRejectsMissingVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{}))
	request := httptest.NewRequest(http.MethodPatch, "/todos/41bb7dd8-20f5-4bd7-93af-2b7163bb7a35", strings.NewReader(`{"completed":true}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "invalid-version")
}

func TestTodosUpdateMapsNotFoundToProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{updateErr: todo.ErrNotFound}))
	request := httptest.NewRequest(http.MethodPatch, "/todos/41bb7dd8-20f5-4bd7-93af-2b7163bb7a35", strings.NewReader(`{"version":1}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), "todo-not-found")
}

func TestTodosUpdateMapsTitleValidationToProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{updateErr: entity.ErrEmptyTodoTitle}))
	request := httptest.NewRequest(http.MethodPatch, "/todos/41bb7dd8-20f5-4bd7-93af-2b7163bb7a35", strings.NewReader(`{"version":1}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "invalid-todo-title")
}

func TestTodosUpdateMapsUnexpectedFailureToProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{updateErr: errors.New("database unavailable")}))
	request := httptest.NewRequest(http.MethodPatch, "/todos/41bb7dd8-20f5-4bd7-93af-2b7163bb7a35", strings.NewReader(`{"version":1}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), "internal-error")
}

func TestTodosDeleteReturnsNotFoundProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{deleteErr: todo.ErrNotFound}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/todos/41bb7dd8-20f5-4bd7-93af-2b7163bb7a35", nil))

	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), "todo-not-found")
}

func TestTodosDeleteReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/todos/41bb7dd8-20f5-4bd7-93af-2b7163bb7a35", nil))

	require.Equal(t, http.StatusNoContent, response.Code)
}

func TestTodosDeleteRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/todos/not-a-uuid", nil))

	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestTodosDeleteMapsUnexpectedFailureToProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.RegisterHandlers(router, handler.NewTodoHandler(stubUseCase{deleteErr: errors.New("database unavailable")}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/todos/41bb7dd8-20f5-4bd7-93af-2b7163bb7a35", nil))

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), "internal-error")
}
