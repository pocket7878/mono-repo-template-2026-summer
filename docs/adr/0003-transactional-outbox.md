# ADR 0003: Transactional Outbox によるイベント生成の整合性

- Status: Accepted
- Date: 2026-07-22

## Context

一つの操作が local database の状態更新と message broker への通知を必要とするとき、二つを通常の手続きとして順番に実行すると不整合が生じる。DB commit 後の publish が失敗すれば後続処理が永続的に欠落する。publish 後の DB commit が失敗すれば、存在しない状態を示す ghost event が配信される。

broker と relational database を一つの transaction に参加させずにこの dual-write 問題を扱い、未送信・再送・恒久失敗を監査可能にする必要がある。

## Decision

1. DB 状態更新を起点に必ず配信すべき event には Transactional Outbox を使う。
2. business data の更新と outbox row の INSERT は同じ DB transaction で commit する。
3. event は UUID、event type、JSON payload、attempt count、status、lease、timestamp を持つ。
4. publisher は `pending` かつ lease 切れの row を小さな batch で claim する。
5. 複数 publisher の競合を避けるため `FOR UPDATE SKIP LOCKED` を使い、claim 時に attempt と lease を更新する。
6. publish 成功時は `published` と `published_at` を記録する。
7. publish 失敗が上限へ達した場合は outbox を `failed` にし、同じ DB transaction で dead-letter row を作る。
8. message envelope の idempotency key は event UUID とする。delivery は at-least-once とし、consumer も冪等でなければならない。
9. published row は監査・再送要件に応じた retention 後に削除する。

## Applicability

次をすべて満たす場合の第一候補とする。

- local DB update と通知の片方だけ成功すると整合性または復旧性を損なう。
- 通知欠落を定期再計算など別経路で自然回復できない。
- 即時性より、再送・監査・障害復旧を優先する。

best-effort telemetry、DB transaction に紐づかない external event、自然回復する cache rebuild、poll latency を許容できない ultra-low-latency flow には自動適用しない。

## Rejected alternatives

### DB commit 後に直接 publish

単純だが、process crash や broker outage の間に event を失う。retry queue 自体へ記録する前に失敗するため復旧根拠がない。

### publish 後に DB commit

DB rollback 後も consumer が event を観測し、ghost event を作る。

### distributed transaction / 2PC

参加可能な基盤、coordinator、障害時運用が必要で、現在の要件には複雑性が過大である。

### periodic full scan だけで後続処理を起動

自然回復する処理には有効だが、個々の要求の監査、低遅延、failure reason の追跡が弱い。

## Consequences

### Positive

- business update と event creation が原子的になる。
- broker outage 中も DB に未送信 event が残り、再開後に送れる。
- pending、published、failed、attempt、last error を監査できる。
- publisher を水平 scale できる。

### Negative

- DB commit と broker delivery の exactly-once は保証しない。publish 成功後、status 更新前に crash すれば重複する。
- poller、lease、retry、DLQ、cleanup、monitoring が必要になる。
- polling interval 分の latency が加わる。
- outbox table が無制限に増えない retention が必要になる。

## Implementation

- schema: `backend/db/migrations/20260722010000_create_job_outbox_events.up.sql`
- worker policy: `backend/internal/usecase/outbox/worker.go`
- PostgreSQL claim/status transition: `backend/internal/repository/outbox.go`
- worker process: `backend/cmd/worker/main.go`
- SQS adapter: `backend/internal/infra/sqs/publisher.go`
- local queue: `backend/docker-compose.yml` と `backend/scripts/localstack/10-create-queues.sh`
- unit test: `backend/internal/usecase/outbox/worker_test.go`
- integration test: `backend/internal/repository/outbox_integration_test.go`
- atomic Todo mutation/outbox insert: `backend/internal/repository/todo.go`
- infrastructure example: `infra/terraform/main.tf` の queue/DLQ

Todo create/update/delete と outbox event は同じ PostgreSQL transaction で保存され、worker は AWS SDK の SQS adapter を通じて queue へ publish する。LocalStack E2E は browser 操作から SQS message 到達までを検証する。consumer inbox/idempotency、automated replay tooling、published event cleanup はまだ実装していない。

## Verification

- 同じ pending row を複数 worker が claim しないこと。
- publish 成功で `published` へ遷移すること。
- publish 失敗で `failed` と dead-letter が同一 transaction で作られること。
- consumer の同一 idempotency key 再処理が二重副作用にならないこと。

```bash
cd backend
go test ./internal/usecase/outbox ./internal/repository ./internal/infra/sqs
make e2e
```

## Operational rules

- oldest pending age、pending/failed count、publish latency、attempt distribution、DLQ count を監視する。
- lease は通常の publish 時間より長くし、process crash 後には有限時間で再claim可能にする。
- replay は元 event ID を維持し、手動の payload 書き換えを避ける。
- dead letter の閲覧、原因修正、再送、監査記録を runbook に定義する。
- payload へ secret や不要な個人情報を含めない。

## Revisit conditions

- polling が database の主要負荷になったとき。
- broker/database が transactional integration を正式に提供するとき。
- event ordering が必要になり、aggregate sequence や partition key が必要になったとき。
