# ADR 0004: API エラーを Problem Details に統一する

- Status: Accepted
- Date: 2026-07-22

## Context

HTTP API の error response が handler ごとに異なると、frontend は `message`、`error`、plain text など複数形式を解釈する必要がある。status code と body の意味が一致しない場合、retry、再入力、再読込の判断も不安定になる。

error response は内部実装の例外表現ではなく、client と共有する安定した契約として定義する必要がある。

## Decision

1. error body は RFC 9457 系の Problem Details 形式を採用する。
2. `Content-Type` は `application/problem+json` とする。
3. 共通 field は `type`、`status`、`title`、`detail` とする。
4. `type` は problem category を示す安定した URI とし、request ごとの message を埋め込まない。
5. `status` は実際の HTTP status と一致させる。
6. `title` は短く安定した分類名、`detail` は安全に公開可能な説明とする。
7. 400 は request syntax/domain validation、404 は resource absence、409 は current state との競合、500 は予期しない内部失敗に使う。
8. 500 response に内部 error、SQL、path、stack trace、secret を含めない。
9. body を `{ "error": ... }` で追加包装しない。
10. machine-readable extension が本当に必要になった場合のみ TypeSpec で追加し、全 client への互換性を評価する。

## Applicability

browser、mobile、CLI、external integration など複数 client が同じ API を使う場合に適用する。streaming protocol や GraphQL のように transport 固有の error model を持つ interface には、その protocol の規約を優先する。

## Rejected alternatives

### handler 固有 JSON

実装は速いが client 分岐と documentation drift を増やすため採用しない。

### HTTP status だけを返す

同じ 400/409 の中で利用者が取るべき行動を説明できない。

### 内部 error message をそのまま detail にする

情報漏えいと、不安定な implementation detail への client dependency を作る。

### 独自 error code を常に必須にする

Problem `type` と HTTP status で足りる段階では二重の分類体系になる。client が分岐を必要とする具体的 use case が生じたときに追加する。

## Consequences

### Positive

- client が一つの parser と表示方針を持てる。
- error response が TypeSpec/OpenAPI の contract review 対象になる。
- 予期しない内部情報の露出を共通方針で防げる。

### Negative

- handler は domain/usecase error を HTTP problem へ明示的に map する必要がある。
- `detail` の文言を UI 表示に使う場合、localization と変更安定性を別途考慮する必要がある。
- validation error の複数 field 表現は標準4 field だけでは不足する場合がある。

## Implementation

- contract: `shared/backend_api/main.tsp` の `Problem` と status-specific model
- generated Go model: `backend/internal/handler/api/api.gen.go`
- handler mapping: `backend/internal/handler/todo.go`
- router fallback: `backend/internal/infra/httpserver/router.go`
- tests: `backend/internal/handler/todo_test.go`

sample の type base URI は `https://reference-todo.dev/problems/` である。実プロダクトでは管理可能で永続的な HTTPS namespace に置き換える。

frontend の `front/app/features/todos/api.ts` は `application/problem+json` を `APIProblemError` として parse し、status と Problem body を保持する。TodoPage は409だけを競合として扱い、その他の update failure と区別する。

## Verification

- contract に各 operation の error status と media type が存在すること。
- response header が `application/problem+json` であること。
- body status と HTTP status が一致すること。
- 400/404/409/500 mapping と、500 の安全な detail を test すること。

```bash
cd backend
go test ./internal/handler ./internal/infra/httpserver
```

## Operational rules

- 新しい handler が独自 error JSON を返すことを禁止する。
- log には内部 error と correlation context を残してよいが、response body と分離する。
- `type` の意味を変更して再利用しない。別の意味には別 URI を使う。
- retryability は status だけで推測できない場合、documented extension または operation semantics で定義する。

## Revisit conditions

- field-level validation error を machine-readable に扱う client が必要になったとき。
- localization、correlation ID、support code を contract に追加する必要が出たとき。
- public API versioning により problem type の互換性管理が必要になったとき。
