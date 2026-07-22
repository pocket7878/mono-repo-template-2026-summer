package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/handler"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/infra/httpserver"
	"github.com/stretchr/testify/require"
)

func TestNewRouterServesHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := httpserver.NewRouter(handler.NewTodoHandler(nil))
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"status":"ok"}`, response.Body.String())
}

func TestNewRouterAnswersCORSPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := httpserver.NewRouter(handler.NewTodoHandler(nil))
	request := httptest.NewRequest(http.MethodOptions, "/todos", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "http://127.0.0.1:5173", response.Header().Get("Access-Control-Allow-Origin"))
}
