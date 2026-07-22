# Generalized Architecture Decisions Implementation Plan

**Goal:** プロダクト固有語を使わず、元リポジトリ相当の情報密度を持つ再利用可能な ADR と Reference Pattern を整備する。

**Architecture:** 実装済み判断は Accepted ADR、未実装の汎用判断は Reference Pattern に分離する。すべての文書から実装・テスト・CI へ辿れるようにし、README を索引にする。

**Tech Stack:** Markdown, Go, TypeSpec, React, GitHub Actions, Terraform

## Global constraints

- 技術的な判断内容を要約しない。
- プロダクト固有語だけを一般化する。
- 未実装事項を Accepted と表現しない。
- 現在のコードを実装状態の正とする。

### Task 1: Accepted ADR を完全版へ改稿する

**Files:**

- Modify: `docs/adr/0001-monorepo-and-contract-source.md`
- Modify: `docs/adr/0002-clean-architecture-quality-gates.md`
- Modify: `docs/adr/0003-transactional-outbox.md`
- Modify: `docs/adr/0004-api-error-model.md`
- Modify: `docs/adr/0005-concurrency-control.md`

- [x] 各 ADR に Applicability、Implementation、Verification、Revisit conditions を追加する。
- [x] 元 ADR の判断、却下案、negative consequence、operational rule を情報量を落とさず一般化する。
- [x] `backend/`、`front/`、`shared/backend_api/`、workflow の実在パスを参照する。
- [x] `rg -n 'Forecast-X|inventory plan|tenant' docs/adr` が意図しない固有語を返さないことを確認する。

### Task 2: 未実装パターンを Reference として完全記述する

**Files:**

- Move/Rewrite: `docs/adr/0006-raw-file-storage.md` to `docs/reference-patterns/raw-file-object-storage.md`
- Create: `docs/reference-patterns/async-worker-queue-consistency.md`

- [x] Status を `Reference` とし、採用済み ADR と区別する。
- [x] object storage について upload、metadata、key、checksum、size limit、IAM、encryption、lifecycle、reconciliation を記載する。
- [x] worker/queue について envelope、at-least-once、冪等性、retryable 分類、backoff、DLQ、replay を記載する。
- [x] 適用条件、非適用条件、却下案、監視項目、障害復旧手順を記載する。

### Task 3: 索引と整合性を完成させる

**Files:**

- Modify: `README.md`
- Modify: `docs/design/reference-architecture.md`

- [x] Accepted ADR と Reference Pattern の違いを説明する。
- [x] すべての設計文書への索引を追加する。
- [x] `rg -n 'Status: Accepted' docs` と実装を照合する。
- [x] `git diff --check`、`make check-gen`、`go run ./cmd/backend-lint ./...` を実行する。
