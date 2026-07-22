# ADR 0003: Transactional Outbox

- Status: Accepted
- Date: 2026-07-22

## Context

DB 更新と broker publish を別々に成功させると、失われたイベントまたは実体のないイベントが発生する。キューは at-least-once 配信なので、再送も前提にする必要がある。

## Decision

業務更新と outbox レコード作成を同一 transaction で commit する。publisher は `FOR UPDATE SKIP LOCKED` で event を claim し、成功時に published、失敗上限時に dead letter へ遷移する。message には event ID を idempotency key として含め、consumer は重複を安全な no-op にする。

## Rejected alternatives

- commit 後の direct publish: publish 失敗時に復旧できない。
- publish 後の DB 更新: ghost event を作る。
- 2PC: この規模には運用複雑性が過大。

## Consequences

再送・監査・障害復旧が可能になる代わりに、poller、retention、DLQ 監視が必要になる。

## Operational rules

- outbox は「DB 状態と通知の原子的整合」が必要な非同期処理にだけ適用する。
- oldest unpublished event、失敗数、DLQ 件数を監視する。
- consumer の外部副作用は必ず冪等にする。
