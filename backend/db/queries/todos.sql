-- name: CreateTodo :exec
INSERT INTO todos (id, title, completed, version, due_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListTodos :many
SELECT id, title, completed, version, due_date, created_at, updated_at
FROM todos
ORDER BY created_at, id;

-- name: FindTodoByID :one
SELECT id, title, completed, version, due_date, created_at, updated_at
FROM todos
WHERE id = $1;

-- name: UpdateTodo :execrows
UPDATE todos
SET title = $1, completed = $2, version = $3, due_date = $4, updated_at = $5
WHERE id = $6 AND version = $7;

-- name: DeleteTodo :execrows
DELETE FROM todos WHERE id = $1;
