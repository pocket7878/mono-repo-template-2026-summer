# Reference Pattern: At-least-once Queue と非同期 Worker の整合性

- Status: Reference
- Date: 2026-07-22

## Context

一般的な queue は、consumer crash、visibility timeout、network timeout、ack failure により同じ message を複数回配信する。順序も partition をまたいで保証されない。worker を「一度だけ、順番通りに実行される」前提で書くと、二重課金、二重通知、重複 row、古い state による上書きが発生する。

## Pattern

1. delivery semantics は at-least-once と明記する。
2. message envelope は `messageId`、`jobId`、`jobType`、`idempotencyKey`、`attempt`、`occurredAt`、`schemaVersion`、payload を持つ。
3. consumer は idempotency key と domain state transition により重複を safe no-op にする。
4. global duplicate ledger を自動採用しない。外部副作用、aggregate update、append operation など、処理の意味に合う冪等化を選ぶ。
5. error を retryable と non-retryable に分類する。
6. timeout、rate limit、一時的 network/database failure は bounded exponential backoff と jitter で retry する。
7. schema violation、unsupported job type、永続的 domain rejection は retry せず DLQ へ移す。
8. maximum attempts 超過も DLQ とし、failure class、last error、安全な payload reference、timestamps を残す。
9. message の delete/ack は business effect の commit 後に行う。
10. ordering が必要な場合は aggregate key で partition し、message に sequence/version を持たせる。queue 全体の順序へ依存しない。

## Idempotency strategies

- conditional state transition: `pending -> processing -> completed` の期待状態を WHERE 条件にする。
- unique constraint: external transaction ID や idempotency key を UNIQUE にする。
- compare-and-swap: aggregate version を条件にする。
- outbox/inbox: event ID の処理済み記録と business update を同じ transaction にする。
- provider key: 支払いや通知 provider が提供する idempotency key を利用する。

処理済み記録だけ先に commit すると、business effect 失敗後の retry を失う。記録と effect の原子性を設計する。

## Retry policy

retry policy は attempt 数だけでなく、最大経過時間、処理 timeout、visibility timeout、backoff 上限を定義する。visibility timeout は通常処理時間より長くし、長時間 job では heartbeat/extension を使う。無限 retry は poison message と障害中の downstream を増幅するため禁止する。

## DLQ and replay

DLQ は墓場ではなく運用 queue である。runbook に以下を含める。

1. message と failure class の確認。
2. secret/個人情報を露出しない diagnostic context の取得。
3. code/config/data の原因修正。
4. replay 前の現在 state と冪等性確認。
5. 元の message ID/idempotency key を維持した bounded replay。
6. replay 結果と operator の監査記録。

payload を手で修正して再送する場合は、新しい message として provenance を残す。

## Rejected alternatives

### exactly-once と仮定する

broker の exactly-once 機能があっても、DB や外部 API の副作用まで自動的に exactly-once にはならない。

### すべてを同じ回数 retry する

permanent error を遅延させ、downstream outage を retry storm で悪化させる。

### catch して成功扱いする

message loss を観測不能にする。

### FIFO queue だけで整合性を担保する

duplicate、consumer crash、外部副作用の原子性は解決しない。

## Observability

- queue depth、oldest message age、processing latency、success/failure rate。
- retry count と error classification。
- DLQ depth と oldest DLQ age。
- duplicate/no-op count、visibility timeout extension、replay result。
- job type、schema version ごとの SLI。

## Security

- message payload に credential、token、不要な個人情報を含めない。
- sensitive input は object reference などへ分離し、worker 側で再認可する。
- queue producer/consumer/DLQ replay の IAM を分離する。
- log に payload 全体を出さない。

## Relationship to Transactional Outbox

Transactional Outbox は producer 側の DB update と event creation を整合させる。この pattern は broker delivery 後の duplicate、retry、DLQ、consumer effect を扱う。どちらか一方で全体の整合性を保証できるわけではない。

## Example assets in this repository

`backend/internal/usecase/outbox/worker.go`、`backend/internal/repository/outbox.go`、`backend/internal/infra/sqs/publisher.go`、`infra/terraform/main.tf` に producer、retry、DLQ の実装例がある。consumer inbox/idempotency、backoff、metrics、automated replay tooling は未実装であるため、consumer 側を含むこの文書全体は Reference Pattern とする。
