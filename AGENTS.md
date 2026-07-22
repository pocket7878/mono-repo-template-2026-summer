# 開発規約

## 実行規約

- 開発コマンドは `nix develop` のシェル内で実行する。Node / npm / npx をホスト環境から直接実行しない。
- PostgreSQL へ接続する前に `backend/.env` があれば必ず読み、既定ポートを推測しない。
- API 契約は `shared/backend_api` の TypeSpec が唯一のソースであり、生成物を直接編集しない。
- backend の依存は `handler -> usecase <- repository -> entity` に従う。handler から repository を直接呼ばない。

## 完了時の検証

- API 契約を変更した場合は `make gen && make check-gen` を実行する。
- backend を変更した場合は `cd backend && gofmt -s -l . && golangci-lint run ./cmd/... ./internal/... && go tool go-arch-lint check && go run ./cmd/backend-lint ./... && ./scripts/run_coverage_gate.sh` を実行する。
- frontend を変更した場合は `cd front && npm run typecheck && npm run lint && npm run fmt:check && npm run coverage:gate` を実行する。
- ユーザーフロー、ルーティング、API 統合、SSR に影響する変更は `make e2e` を実行する。
