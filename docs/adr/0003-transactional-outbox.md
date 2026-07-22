# ADR 0003: Transactional Outbox と DLQ

- Status: Accepted
- Date: 2026-07-22

## Context

業務データの更新と broker publish を別の操作にすると、DB だけ更新される lost event と、実体のない ghost event が発生する。

## Decision

非同期処理の要求は業務更新と同じトランザクションで `job_outbox_events` に記録する。Worker は未送信イベントを `FOR UPDATE SKIP LOCKED` で claim し、成功時は published、最大試行回数に到達した失敗は `job_dead_letters` に移す。

## Consequences

at-least-once delivery を前提に consumer は冪等にする必要がある。一方、未送信・失敗イベントを DB から再送・監査できる。

## Operational Rules

Worker の attempt を envelope の idempotency key として downstream に渡す。DLQ を再送する前に failure reason と payload を確認する。
