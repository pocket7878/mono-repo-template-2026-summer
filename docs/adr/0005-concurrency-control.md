# ADR 0005: 更新競合を optimistic locking で検出する

- Status: Accepted
- Date: 2026-07-22

## Context

同じ resource を複数の browser tab、利用者、worker が読み取り、異なる時点の値を更新することがある。無条件 UPDATE は後から到着した request を成功させ、先に保存された変更を通知なく失わせる。

この sample では Todo が aggregate の最小例であり、title、completed、due date を一つの version で保護する。目的は同時実行を直列化することではなく、stale write を検出して呼び出し元へ戻すことである。

## Decision

1. 更新可能な aggregate は正の整数 `version` を持つ。
2. create 時の version は 1 とする。
3. read response に current version を含め、update request は client が読んだ expected version を送る。
4. usecase は next entity を `expectedVersion + 1` で構築する。
5. repository の UPDATE は primary key と expected version の両方を WHERE 条件にする。
6. affected row が 0 の場合は version conflict として扱う。ただし resource absence と conflict を厳密に区別する必要がある API では、追加 read または DB returning strategy を設計する。
7. conflict は HTTP 409 Problem Details に map する。
8. client は成功とみなさず、最新 state を再取得し、利用者に再適用または再編集を促す。
9. last-write-wins への暗黙 fallback を行わない。

## Applicability

read-modify-write があり、更新衝突は頻繁ではないが lost update を許容できない resource に適用する。高頻度 counter、append-only log、在庫引当のように別の atomic operation や強い serialization が必要な domain には、その不変条件に合う方式を選ぶ。

## Rejected alternatives

### last-write-wins

単純だが、利用者が読んでいない変更を黙って消すため採用しない。

### transaction 中の悲観 lock を UI 操作全体で保持

request 間で lock を保持できず、長時間 lock は throughput と障害復旧を悪化させる。

### timestamp 比較

精度、clock、同値、意味が曖昧になる。単調増加する整数 version の方が契約として明確である。

### frontend だけで競合を防ぐ

複数 client、worker、直接 API 呼び出しを防げず、最終防衛線にならない。

## Consequences

### Positive

- lost update を deterministic に検出できる。
- row lock を長時間保持しない。
- conflict behavior を API contract と integration test で表現できる。

### Negative

- request/response と永続化 schema に version が伝播する。
- client に conflict UX が必要になる。
- 複数 aggregate を一括更新する場合は version だけで transaction-level invariant を表現できない。

## Implementation

- schema: `backend/db/migrations/20260722000000_create_todos.up.sql`
- contract: `shared/backend_api/main.tsp` の Todo/UpdateTodoRequest version
- entity: `backend/internal/entity/todo.go`
- usecase: `backend/internal/usecase/todo/usecase.go`
- conditional UPDATE: `backend/internal/repository/todo.go`
- error mapping: `backend/internal/handler/todo.go`
- integration test: `backend/internal/repository/todo_integration_test.go`
- API test: `backend/internal/handler/todo_test.go`

frontend は update request に version を送り、`APIProblemError.status === 409` のときだけ競合回復メッセージを表示する。conditional UPDATE の affected row 0 は、同時 delete も 409 として扱う。

## Verification

1. version 1 の resource を expected version 1 で更新すると version 2 になる。
2. 同じ expected version 1 でもう一度更新すると repository が conflict を返す。
3. handler が 409 と `application/problem+json` を返す。
4. successful response が新しい version を返す。

```bash
cd backend
go test ./internal/usecase/todo ./internal/repository ./internal/handler
```

## Operational rules

- version column を直接固定値へ戻さない。
- bulk/import/worker update も同じ invariant を守るか、別 ADR で atomicity を定義する。
- 409 を自動無限 retry しない。merge strategy が明確な operation だけ bounded retry を許可する。
- conflict rate の急増は client bug、hot resource、retry storm の兆候として観測する。

## Revisit conditions

- conflict が通常系となり、再編集 UX では扱えないとき。
- aggregate が複数 row/table に拡大し、単一 version の境界を再定義するとき。
- compare-and-swap では表現できない順序保証や reservation が必要になったとき。
