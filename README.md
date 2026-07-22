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
make setup
make gen
make check
```

`direnv` を使う場合は、初回だけ `direnv allow` でこのリポジトリの `.envrc` を許可します。以後はディレクトリへ移動するだけで Nix 開発環境が有効になります。使わない場合は、`nix develop` のシェル内で以降のコマンドを実行してください。

`make check` は生成同期、backend coverage gate、frontend テストを実行します。主要フローは `make e2e` で Playwright により検証します。`flake.lock` はリポジトリで管理し、開発環境を固定します。

### 3. ローカルで起動する

Docker Desktop を起動してから、初回のみ PostgreSQL とスキーマを準備します。`make db-migrate` は未適用の初期マイグレーションを適用するコマンドであり、同じデータベースへ繰り返し実行しません。

```bash
make db-migrate
```

開発サーバーは別々のターミナルで起動します。

```bash
# terminal 1
cd backend && DATABASE_URL='postgres://postgres:postgres@localhost:15432/reference_todo?sslmode=disable' go run ./cmd/server

# terminal 2
cd front && npm run dev
```

画面は `http://127.0.0.1:5173/todos`、API のヘルスチェックは `http://127.0.0.1:8080/health` です。

詳しい設計は [docs/design/reference-architecture.md](docs/design/reference-architecture.md) を参照してください。
