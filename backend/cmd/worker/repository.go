package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/repository"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/outbox"
)

func newOutboxRepository(pool *pgxpool.Pool) outbox.Repository {
	return repository.NewOutboxRepository(pool)
}
