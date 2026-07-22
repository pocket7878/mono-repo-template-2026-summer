# Reference Todo Architecture

Go、React、TypeSpec、Terraform を組み合わせた、実行可能なフルスタック・モノレポのリファレンス実装です。
ToDo を題材に、契約駆動開発、Clean Architecture、アーキテクチャ lint、品質ゲート、非同期処理、CI/CD を再現します。

## 構成

- `backend/`: Go API / Worker、PostgreSQL、Clean Architecture
- `front/`: React Router による ToDo UI
- `shared/backend_api/`: TypeSpec を唯一の API 契約として管理
- `infra/terraform/`: AWS 構成の実装例

## セットアップ

### 1. 前提条件

- macOS または Linux
- Docker Desktop（ローカル PostgreSQL と E2E テストに使用）
- Nix（flakes が利用可能なこと）
- 任意: direnv と nix-direnv（入退室時の開発環境自動有効化に使用）

Nix の導入方法、flakes の有効化方法、`direnv`/`nix-direnv` のインストール方法は、OS や既存の Nix 配布形態によって異なります。このリポジトリでは特定の導入方法を規定しません。利用中の配布形態に対応した [Nix 公式ドキュメント](https://nixos.org/manual/nix/stable/) と [nix-direnv のドキュメント](https://github.com/nix-community/nix-direnv) に従って準備してください。

### 2. リポジトリを初期化する

```bash
git clone https://github.com/pocket7878/mono-repo-template-2026-summer.git
cd mono-repo-template-2026-summer
direnv allow
make setup-local-worktree-env
make setup
make gen
make check
```

`direnv` を使う場合は、初回だけ `direnv allow` でこのリポジトリの `.envrc` を許可します。以後はディレクトリへ移動するだけで Nix 開発環境が有効になります。使わない場合は、`nix develop` のシェル内で以降のコマンドを実行してください。

`make setup-local-worktree-env` は repository の物理 path から Compose project と PostgreSQL、LocalStack、backend、frontend の port を決め、Git 管理外の `backend/.env` に保存します。これにより複数 worktree を並行起動できます。既存ファイルを作り直す場合だけ `./scripts/setup-local-worktree-env.sh --force` を使用してください。

`make check` は生成同期、backend custom lint、backend coverage gate、frontend unit test を実行します。format、一般 lint、architecture lint、typecheck は各 CI と `AGENTS.md` のコマンドを参照してください。主要フローは `make e2e` で Playwright により検証します。`flake.lock` はリポジトリで管理し、開発環境を固定します。

DB query は `backend/db/queries` を編集し、`make gen` で sqlc の型付きコードを更新します。生成先 `backend/internal/repository/internal/db` は直接編集しません。`make check-gen` は API と DB の生成同期をまとめて検証します。

### 3. ローカルで起動する

Docker Desktop を起動してから、初回のみ PostgreSQL とスキーマを準備します。`make db-migrate` は未適用の初期マイグレーションを適用するコマンドであり、同じデータベースへ繰り返し実行しません。

```bash
make db-migrate
```

開発サーバーは別々のターミナルで起動します。

```bash
# terminal 1
cd backend && set -a && . ./.env && set +a && go run ./cmd/server

# terminal 2
cd backend && set -a && . ./.env && set +a
cd ../front && npm run dev -- --port "${FRONTEND_PORT}"

# terminal 3
cd backend && set -a && . ./.env && set +a && go run ./cmd/worker
```

実際の URL は `backend/.env` の `FRONTEND_PORT` と `BACKEND_ADDR` で確認できます。単一 checkout で固定 port を使いたい場合は `.env.example` の値を `backend/.env` にコピーしてください。

## 設計ドキュメント

- [リファレンス全体の設計](docs/design/reference-architecture.md)
- [ADR の執筆方針](docs/design/adr-authoring-policy.md)
- [Monorepo と TypeSpec 契約単一ソース](docs/adr/0001-monorepo-and-contract-source.md)
- [依存方向と機械的品質ゲート](docs/adr/0002-clean-architecture-quality-gates.md)
- [Transactional Outbox](docs/adr/0003-transactional-outbox.md)
- [Problem Details](docs/adr/0004-api-error-model.md)
- [optimistic locking](docs/adr/0005-concurrency-control.md)

未実装部分を含む再利用可能な設計パターンは、採用済み ADR と区別しています。

- [非同期入力ファイルと object storage](docs/reference-patterns/raw-file-object-storage.md)
- [at-least-once queue と worker の整合性](docs/reference-patterns/async-worker-queue-consistency.md)

## License

[MIT License](LICENSE)
