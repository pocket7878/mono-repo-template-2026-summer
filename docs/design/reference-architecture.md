# リファレンス・アーキテクチャ設計

- Status: Proposed
- Date: 2026-07-22
- Owner: Masato

## 目的

このリポジトリは、Go、React、TypeSpec、Terraform を使うフルスタック・モノレポの、実行可能なリファレンス実装とする。
特定の業務システムの雛形ではなく、今後の新規プロジェクトや異なる組織で再利用できる設計・品質・運用の基準点を提供する。

小さな ToDo アプリケーションを題材にする。題材そのものではなく、実務上の設計判断を最小規模で再現し、コード、設定、テスト、CI でその判断を強制できることを成功条件とする。

## スコープ

含めるもの:

- Go バックエンド、React フロントエンド、TypeSpec API 契約、Terraform を同一リポジトリで管理するモノレポ
- ToDo の作成、一覧、更新、完了、削除、タグ・期限検索
- CSV 一括インポートを題材にした非同期ジョブ、Transactional Outbox、Worker、再試行、DLQ と運用手順
- 楽観ロックによる更新競合と API エラー契約
- 生成物同期、静的解析、アーキテクチャ検証、ユニット・統合・E2E テスト、CI
- worktree ごとのローカル環境分離

含めないもの:

- 特定の会社、ドメイン、認証事業者、AWS アカウント、環境名への依存
- 題材を超えるプロダクト機能や汎用 SaaS 機能
- 設計を強制できない一般論だけの文書集

## リポジトリ構造

```text
backend/                 # Go: HTTP API と非同期 Worker
  cmd/server/            # 起動・DI・ライフサイクル管理
  cmd/worker/            # 非同期ジョブ実行の入口
  internal/entity/       # ドメイン型と不変条件
  internal/usecase/      # ビジネスロジックと Repository interface
  internal/repository/   # PostgreSQL、Queue、Outbox の実装
  internal/handler/      # HTTP 入出力、認証、Problem Details
  db/                    # migration、query、生成コード
  scripts/               # テスト DB、coverage、migration 検証
front/                   # React Router による UI
  app/features/          # ToDo・インポート等の feature
  app/routes/            # 画面・ルーティング境界
  app/lib/backend-api/   # 契約から生成した API 型・クライアント
  e2e/                   # Playwright の主要フロー
shared/backend_api/      # TypeSpec を唯一の API 契約として管理
infra/terraform/         # AWS 構成の実装例
scripts/                 # 生成同期、ローカル E2E、worktree 環境分離
docs/adr/                # 採用した設計判断
docs/design/             # リファレンス自体の設計
```

## アーキテクチャと依存方向

バックエンドの依存は `handler -> usecase <- repository -> entity` とする。

- `entity` はドメインの不変条件を型とコンストラクタで表現し、プロジェクト内の他層へ依存しない。
- `usecase` は業務処理と、必要な Repository interface を定義する。
- `repository` は interface を実装し、DB 復元に限り entity の `Reconstruct*` を使う。
- `handler` は HTTP 入出力を変換し、usecase のみを呼び出す。repository を直接呼び出さない。
- DB 生成コードは `repository/internal/db` に閉じ込め、他層から参照できないようにする。

この配置は `go-arch-lint` で検証する。設計規約を README の注意書きに留めず、違反した変更を CI で失敗させる。

## 契約とデータフロー

`shared/backend_api` の TypeSpec を API 契約の唯一のソースとする。

```text
TypeSpec -> OpenAPI -> Go handler 生成物
                  -> TypeScript client / 型
```

契約変更では、契約、生成物、backend、frontend を同一変更セットで更新する。CI は生成をやり直し、未コミットの差分が発生した場合に失敗させる。

通常の更新は HTTP handler から usecase を経て PostgreSQL に保存する。ToDo 更新はバージョン番号を条件に含め、競合時には `409 Conflict` を返す。

CSV インポートは HTTP リクエスト内で実データを処理しない。ジョブと Outbox event を同一トランザクションで保存し、Worker が Queue 経由で処理する。失敗は再試行後に DLQ へ送り、Runbook で確認・再実行できる。

## エラー処理

API エラーは API 契約で定義し、入力検証、認可、未存在、競合、非同期失敗を一貫した Problem Details 形式に正規化する。

- 入力不正: `400 Bad Request`
- 認証・認可不足: `401 Unauthorized` / `403 Forbidden`
- リソース未存在: `404 Not Found`
- 楽観ロック競合: `409 Conflict`
- 予期しない障害: `500 Internal Server Error`。内部詳細をクライアントへ露出しない

Frontend は契約型を通じてこれらを扱い、利用者が再試行・修正・更新を選べる表示にする。

## 品質ゲート

以下は現時点の標準設定として固定する。設定値は参考例ではなく、リファレンス実装が守る合格基準である。

| 領域 | 基準 | 強制方法 |
|---|---|---|
| Go 書式 | `gofmt -s` | ローカル・CI |
| Go 静的解析 | `golangci-lint` + `go vet` | ローカル・CI |
| Go 複雑度 | `gocognit` の `min-complexity: 15` | `golangci-lint` |
| Go の層 | `go-arch-lint` の依存ルール | ローカル・CI |
| Go 固有規約 | `backend-lint` | ローカル・CI |
| Backend coverage | scoped coverage `85.0%` 以上 | coverage gate |
| Frontend 静的品質 | typecheck、ESLint、Prettier | ローカル・CI |
| Frontend coverage | scoped coverage `60.0%` 以上 | coverage gate |
| 高リスク Frontend 変更 | 変更した route / server API は line coverage `80.0%` 以上 | coverage gate |
| API 生成物 | TypeSpec / OpenAPI / client の差分なし | 生成同期チェック |
| DB migration | 命名 lint と一般アプリ権限での `up -> down -> up` | CI |
| E2E | 主要フローを Playwright で検証 | backend、front、契約変更時の CI |

Backend scoped coverage からは `cmd/**`、`internal/handler/api/**`、`internal/repository/internal/db/**`、`internal/lint/**` を除外する。これは生成物、起動配線、Lint 実装によって業務ロジックの品質指標が歪むことを避けるためであり、任意の除外を許可するものではない。

## 開発・運用

- Node、Go、Terraform、静的解析ツールは `flake.nix` の `nix develop` で固定する。`flake.lock` は必ずコミットする。
- Makefile をリポジトリ横断の入口とし、生成、DB、テスト、開発サーバーを一貫した名前で提供する。
- Docker Compose の project 名とポートは worktree のパスから決定し、並行開発で衝突しないようにする。
- CI は path filter で不要なジョブを避ける。一方、backend、front、API 契約のいずれかに影響する変更では E2E を必ず実行する。
- Terraform はアプリケーションと分離して管理するが、デプロイに必要な構成と CI の検証例を含める。

## ADR

実装と同時に、少なくとも以下を ADR として記録する。

1. Monorepo と TypeSpec 契約単一ソース
2. Clean Architecture と `go-arch-lint` による境界強制
3. テスト戦略と品質ゲート
4. API エラーモデル
5. Transactional Outbox とジョブ失敗の扱い
6. 楽観ロックによる ToDo 更新競合の扱い

各 ADR は、Context、Decision、Consequences、Operational Rules を含める。代替案を選べない技術前提ではなく、判断理由と運用上の約束を残す。

## 検証と完了条件

リファレンス実装の完了は、次をすべて満たすこととする。

1. 新しい開発者が README の手順だけでローカル起動と ToDo の主要操作を実行できる。
2. 意図的な層違反、生成物の更新漏れ、複雑度超過、カバレッジ不足が各ゲートで失敗する。
3. CSV インポート成功、失敗、再試行、DLQ 確認をローカルまたは E2E で再現できる。
4. 更新競合が `409 Conflict` となり、Frontend で回復行動を提示できる。
5. CI が backend、front、契約、infra の責務に沿って検証を実行する。

## 実装順序

1. リポジトリ基盤、ツールチェーン、Makefile、ローカル環境分離
2. TypeSpec 契約と生成パイプライン
3. ToDo の最小 CRUD を Clean Architecture で実装
4. `go-arch-lint`、静的解析、coverage gate、生成同期を有効化
5. Frontend と主要 E2E を実装
6. インポート、Outbox、Worker、DLQ、Runbook を追加
7. Terraform、CI、ADR を完成させる
