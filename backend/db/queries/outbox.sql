-- name: InsertOutboxEvent :exec
INSERT INTO job_outbox_events (id, event_type, payload)
VALUES ($1, $2, $3);

-- name: ClaimPendingOutboxEvents :many
UPDATE job_outbox_events
SET attempt_count = attempt_count + 1, lease_expires_at = sqlc.arg(lease_expires_at), updated_at = sqlc.arg(now)
WHERE id IN (
  SELECT id FROM job_outbox_events
  WHERE status = 'pending' AND (lease_expires_at IS NULL OR lease_expires_at < sqlc.arg(now))
  ORDER BY created_at, id
  FOR UPDATE SKIP LOCKED
  LIMIT sqlc.arg(batch_limit)
)
RETURNING id, event_type, payload, attempt_count;

-- name: MarkOutboxPublished :exec
UPDATE job_outbox_events
SET status = 'published', published_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: LockOutboxEvent :one
SELECT event_type, payload
FROM job_outbox_events
WHERE id = $1
FOR UPDATE;

-- name: MarkOutboxFailed :exec
UPDATE job_outbox_events
SET status = 'failed', last_error = $2, updated_at = NOW()
WHERE id = $1;

-- name: InsertDeadLetter :exec
INSERT INTO job_dead_letters (outbox_event_id, event_type, payload, failure_reason)
VALUES ($1, $2, $3, $4);
