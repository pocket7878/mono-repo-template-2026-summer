# SQS DLQ investigation and replay

## Trigger

`ApproximateNumberOfMessagesVisible > 0` on the outbox DLQ means automatic delivery attempts were exhausted. Treat the payload as potentially sensitive and do not paste it into chat, issue, or log output.

## Investigation

1. Record alarm time, queue, environment, and deploy version.
2. Read one message without deleting it and inspect `idempotencyKey`, `eventType`, `attempt`, and safe metadata.
3. Correlate the event ID with `job_outbox_events` and `job_dead_letters`.
4. Classify the failure as transient infrastructure, invalid message/schema, application defect, or downstream rejection.
5. Confirm that replaying the same idempotency key is a safe no-op or resumes an incomplete operation.

## Replay

1. Fix code, configuration, permission, or data before replay.
2. Preserve the original message body and idempotency key.
3. Move a bounded number of messages to the source queue.
4. Observe source queue depth, processing failures, duplicate/no-op count, and DLQ depth.
5. Stop replay if the same failure recurs.
6. Record operator, time, message IDs, cause, remediation, and result.

Do not bulk redrive an unbounded DLQ. If a payload must be changed, create a new event with provenance instead of silently modifying the failed message.
