# ADR 0002: 依存方向と機械的品質ゲート

- Status: Accepted
- Date: 2026-07-22

## Context

ディレクトリを `handler`、`usecase`、`repository`、`entity` に分けるだけでは architecture は維持できない。締切や自動生成の都合で handler が DB を直接呼ぶ、repository が domain validation を迂回して entity を組み立てる、といった逸脱はコンパイル可能である。

同様に、format、complexity、coverage、generated-file synchronization を個人判断にすると、レビュー前の品質が変更者ごとに変わる。規約は文章だけでなく、ローカルと CI の同じコマンドで失敗させる必要がある。

## Decision

1. backend の主要な処理方向を `handler -> usecase <- repository` とし、domain type は `entity` に置く。handler、usecase、repository は責務に応じて entity type を参照できる。
2. handler から repository を直接呼ばない。usecase は repository interface に依存し、具体実装を知らない。
3. `go-arch-lint` で package 間依存を検査する。
4. repository が永続化済み値から entity を復元するときは `entity.Reconstruct*` を使用する。
5. repository の本番コードで `entity.New*`、entity composite literal、`new(entity.Type)` を使うことを custom Go analyzer で禁止する。test file は fixture 構築のため対象外とする。
6. Go の cognitive complexity は `gocognit` 15 を下限基準とする。
7. backend scoped coverage は 85.0% 以上とする。`cmd/**`、generated API、repository generated DB code、custom lint 自身など、指標を歪める glue/generated code は明示的に除外する。
8. frontend scoped coverage は 60.0% 以上とし、高リスク server-side path を導入した場合は 80.0% 以上とする。
9. 静的検証、unit/integration、E2E の順で feedback を得る。

## Applicability

複数の adapter を持つ application、domain rule を framework/DB から分離したい application、複数人または AI が継続変更する repository に適用する。単一 script まで同じ layer 数へ分割することは求めない。

`Reconstruct*` は DB など信頼済み persisted state の復元専用である。外部 request、command、通常の application flow では validation を持つ `New*` を使う。

## Rejected alternatives

### code review だけで守る

機械的に判定できる違反へ人間の注意力を使い、見落としも発生するため採用しない。

### coverage 100% を要求する

generated code、起動配線、defensive branch のために価値の低い test が増える。重要な business behavior と境界の test を優先し、scoped threshold を使う。

### package-private constructor を前提にする

Go には同一 module 内の特定 package だけへ公開する仕組みがない。reflection や重複 struct を使わず、命名規約と analyzer で意図を強制する。

## Consequences

### Positive

- architecture drift を PR 前に検出できる。
- domain validation と DB 復元の意図が code 上で判別できる。
- coverage と complexity の基準が変更者に依存しない。
- AI が変更する際も、禁止事項が実行可能な feedback になる。

### Negative

- analyzer、allowlist、coverage scope の保守が必要になる。
- threshold は品質そのものではなく、設計レビューを代替しない。
- layer をまたぐ正当な変更でも、interface と adapter の追加が必要になる。

## Implementation

- architecture rule: `backend/.go-arch-lint.yml`
- lint configuration: `backend/.golangci.yml`
- Reconstruct analyzer: `backend/internal/lint/analyzers/repositoryentityreconstruct/analyzer.go`
- custom lint entrypoint: `backend/cmd/backend-lint/main.go`
- coverage scope: `backend/cmd/coverage-gate/main.go`
- SQL source and generator: `backend/db/queries`、`backend/sqlc.yaml`、`backend/internal/repository/internal/db`
- backend gate: `backend/scripts/run_coverage_gate.sh`
- frontend gate: `front/scripts/coverage_gate.py`
- CI: `.github/workflows/backend-ci.yml` と `front-ci.yml`

## Verification

```bash
cd backend
gofmt -s -l .
golangci-lint run ./cmd/... ./internal/...
go tool go-arch-lint check
go run ./cmd/backend-lint ./...
sqlc vet
sqlc generate
./scripts/run_coverage_gate.sh

cd ../front
npm run typecheck
npm run lint
npm run fmt:check
npm run coverage:gate
```

root の `make check` は生成同期、custom backend lint、backend coverage、frontend unit test の短い共通入口である。これは gofmt、golangci-lint、go-arch-lint、frontend typecheck/format、E2E を包含しないため、変更範囲に応じて上記コマンドまたは CI を併用する。主要 user flow または API integration を変更した場合は `make e2e` も実行する。

## Operational rules

- lint disable、coverage exclude、threshold 引き下げは理由と影響範囲をレビューする。
- `Reconstruct*` は validation bypass であることを comment と命名で明示する。
- flaky test は再実行で隠さず、原因または隔離期限を記録する。
- 新規 layer/component は architecture config と CI path filter を同時に更新する。

## Revisit conditions

- repository pattern を廃止する設計へ移行したとき。
- coverage scope が実際の risk と乖離した証拠が得られたとき。
- CI 時間が開発 feedback を阻害し、test partitioning が必要になったとき。
