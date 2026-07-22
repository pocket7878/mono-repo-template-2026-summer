package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/handler"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/infra/httpserver"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/repository"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/todo"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, requiredEnv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	useCase := todo.NewTodoUseCase(repository.NewTodoRepository(pool), func() time.Time { return time.Now().UTC() })
	router := httpserver.NewRouter(handler.NewTodoHandler(useCase))
	server := &http.Server{
		Addr:              envOrDefault("BACKEND_ADDR", ":8080"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
