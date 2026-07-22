# リファレンス・アーキテクチャ基盤 実装計画

**Goal:** Go、React、TypeSpec、Terraform のモノレポとして、契約駆動・Clean Architecture・品質ゲートを実際に実行できる ToDo リファレンスの第1リリースを構築する。

**Architecture:** API 契約は `shared/backend_api` の TypeSpec を唯一のソースとし、OpenAPI を経由して Go と TypeScript の生成物を更新する。バックエンドは `handler -> usecase <- repository -> entity` の依存方向を `go-arch-lint` で固定し、最初の縦切りとして ToDo CRUD と楽観ロックを実装する。

**Tech Stack:** Go 1.25、Gin、PostgreSQL 17、sqlc、TypeSpec、OpenAPI、React 19、React Router v7、TypeScript、Vitest、Playwright、Terraform、Docker Compose、mise、GitHub Actions。

## 全体制約

- TypeSpec を API 契約の唯一のソースとする。生成物を手編集しない。
- Go の依存は `handler -> usecase <- repository -> entity` に限定し、`go-arch-lint` で検証する。
- `gocognit` の `min-complexity` は `15` とする。
- backend scoped coverage は `85.0%` 以上とする。`cmd/**`、`internal/handler/api/**`、`internal/repository/internal/db/**`、`internal/lint/**` だけを除外する。
- frontend scoped coverage は `60.0%` 以上、変更した route / server API は line coverage `80.0%` 以上とする。
- Node / npm / npx は `mise exec node --` 経由で実行する。
- CI では書式、静的解析、生成同期、migration、テスト、coverage、E2E を検証する。
- コミットする場合は Conventional Commits の日本語メッセージを使い、タスクごとに対象ファイルだけを stage する。

## リリース分割

この設計は独立して検証できる三つのリリースに分ける。

1. **本計画: 基盤と ToDo の同期 CRUD** — モノレポ、契約、層、Lint、coverage、React UI、E2E。
2. **後続計画: 非同期インポート** — CSV import、Transactional Outbox、Worker、再試行、DLQ、Runbook。
3. **後続計画: 配備リファレンス** — Terraform、AWS リソース、環境別デプロイ、CI/CD。

第1リリースは、品質ゲートが実装済みの ToDo CRUD をローカルと CI で再現できる状態で完了とする。後続リリースのための `cmd/worker`、Queue、AWS リソースはこの段階では作らない。

## ファイル構成

| パス | 責務 |
|---|---|
| `README.md` | 初期セットアップ、生成、検証、設計の入口 |
| `AGENTS.md` | 人間・エージェント共通の実行規約と必須ゲート |
| `Makefile` | リポジトリ横断コマンド |
| `.github/workflows/*.yml` | path-filter 付きの CI |
| `shared/backend_api/main.tsp` | ToDo API の契約 |
| `shared/backend_api/tspconfig.yaml` | OpenAPI 生成設定 |
| `backend/internal/entity/todo.go` | ToDo の不変条件と再構築 |
| `backend/internal/usecase/todo/*.go` | ユースケースと repository interface |
| `backend/internal/repository/todo.go` | PostgreSQL 実装 |
| `backend/internal/handler/todo.go` | HTTP 入出力と Problem Details への変換 |
| `backend/.go-arch-lint.yml` | 層依存の強制 |
| `backend/.golangci.yml` | gocognit を含む静的解析 |
| `backend/cmd/coverage-gate/*.go` | scoped coverage 計算 |
| `front/app/features/todos/*` | ToDo UI と API client 呼び出し |
| `front/app/routes/todos.tsx` | 一覧・作成・更新画面 |
| `front/scripts/coverage_gate.py` | frontend coverage gate |
| `front/e2e/todos.e2e.spec.ts` | 主要ユーザーフロー |
| `docs/adr/*.md` | 構成・契約・品質の判断記録 |

---

### Task 1: リポジトリの実行入口と開発規約を作る

**Files:**

- Create: `README.md`
- Create: `AGENTS.md`
- Create: `Makefile`
- Create: `.gitignore`
- Create: `mise.toml`
- Create: `docs/adr/0001-monorepo-and-contract-source.md`

**Produces:** `make gen`、`make check-gen`、`make backend-test`、`make frontend-test`、`make e2e` が以降のタスクで参照できる共通コマンドになる。

- [ ] **Step 1: README の受け入れ条件を先に書く**

  `README.md` に、以下の完了操作をコードブロックで記述する。

  ```bash
  mise trust
  make setup
  make gen
  make check
  make dev-up
  ```

  `make check` が format、lint、unit / integration test、coverage、生成同期を実行し、`make e2e` が Playwright を実行することも明記する。

- [ ] **Step 2: 空の Makefile 実行が失敗することを確認する**

  Run: `make check`

  Expected: `make: *** No rule to make target 'check'.  Stop.`

- [ ] **Step 3: Makefile と mise を実装する**

  `Makefile` に次のターゲットを実装する。

  ```make
  .PHONY: setup gen check-gen backend-test frontend-test e2e check

  setup:
	cd shared/backend_api && mise exec node -- npm ci
	cd front && mise exec node -- npm ci
	cd backend && go mod download

  gen:
	cd shared/backend_api && mise exec node -- npx tsp compile .
	cd backend && go generate ./internal/handler/api
	cd front && bash ./tools/gen-api-client.sh

  check-gen:
	python3 scripts/check_generated_file_sync.py -- make gen

  backend-test:
	cd backend && ./scripts/run_coverage_gate.sh

  frontend-test:
	cd front && mise exec node -- npm run test

  e2e:
	./scripts/run-e2e-local.sh

  check: check-gen backend-test frontend-test
  ```

  `mise.toml` は Node 20 と Go 1.25 を固定する。

- [ ] **Step 4: 最小の規約と ADR を書く**

  `AGENTS.md` に `mise exec node --` 規約、DB 接続前に `backend/.env` を読む規約、各変更に必要なゲートを記載する。ADR は Context、Decision、Consequences、Operational Rules を使い、`backend/`、`front/`、`shared/backend_api/`、`infra/terraform/` を一つのリポジトリに置く理由と契約単一ソースを記録する。

- [ ] **Step 5: Makefile の構文を検証する**

  Run: `make -n check`

  Expected: `check-gen`、`backend-test`、`frontend-test` のコマンドが順に表示される。

- [ ] **Step 6: コミットする**

  ```bash
  git add README.md AGENTS.md Makefile .gitignore mise.toml docs/adr/0001-monorepo-and-contract-source.md
  git commit -m "build: リファレンス基盤を追加"
  ```

### Task 2: TypeSpec 契約と生成同期を構築する

**Files:**

- Create: `shared/backend_api/package.json`
- Create: `shared/backend_api/tspconfig.yaml`
- Create: `shared/backend_api/main.tsp`
- Create: `shared/backend_api/models/todo.tsp`
- Create: `scripts/check_generated_file_sync.py`
- Create: `backend/internal/handler/api/generate.go`
- Create: `front/tools/gen-api-client.sh`
- Test: `shared/backend_api/main.tsp`

**Consumes:** Task 1 の `make gen` と `make check-gen`。

**Produces:** `Todo`, `CreateTodoRequest`, `UpdateTodoRequest`, `Problem` の OpenAPI 契約と、backend / frontend の再生成可能な出力。

- [ ] **Step 1: 失敗する契約コンパイルを作る**

  `shared/backend_api/main.tsp` に未定義の `TodoService` を参照する最小プログラムを置く。

  Run: `cd shared/backend_api && mise exec node -- npx tsp compile .`

  Expected: `TodoService` が見つからない TypeSpec エラー。

- [ ] **Step 2: ToDo API 契約を定義する**

  `models/todo.tsp` に次の型を定義する。

  ```typespec
  model Todo {
    id: string;
    title: string;
    completed: boolean;
    version: int32;
    dueDate?: utcDateTime;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
  }

  model CreateTodoRequest { title: string; dueDate?: utcDateTime; }
  model UpdateTodoRequest { title?: string; completed?: boolean; dueDate?: utcDateTime; version: int32; }
  model Problem { type: string; title: string; status: int32; detail?: string; }
  ```

  `main.tsp` には `GET /todos`、`POST /todos`、`PATCH /todos/{todoId}`、`DELETE /todos/{todoId}` を置き、更新操作の `409` 応答を `Problem` として宣言する。

- [ ] **Step 3: OpenAPI の生成成功を確認する**

  Run: `cd shared/backend_api && mise exec node -- npx tsp compile . && test -f tsp-output/@typespec/openapi3/openapi.yaml`

  Expected: exit code 0。

- [ ] **Step 4: backend / frontend の生成入口を実装する**

  `backend/internal/handler/api/generate.go` は `//go:generate` で OpenAPI から Gin 用 handler 型を生成する。`front/tools/gen-api-client.sh` は `openapi-typescript` を実行し、`front/app/lib/backend-api/openapi/types.ts` を上書きする。

- [ ] **Step 5: 生成同期チェックを実装する**

  `scripts/check_generated_file_sync.py` は `make gen` 実行後に `git diff --exit-code -- shared/backend_api/tsp-output backend/internal/handler/api front/app/lib/backend-api/openapi/types.ts` を呼び、差分があれば exit code 1 にする。

- [ ] **Step 6: 生成同期の成功と失敗を確認する**

  Run: `make gen && make check-gen`

  Expected: exit code 0。

  `main.tsp` の `Todo.title` を一時的に `description` へ変更し、生成を実行せずに `make check-gen` を実行する。

  Expected: 生成物差分を表示して exit code 1。

- [ ] **Step 7: コミットする**

  ```bash
  git add shared/backend_api scripts/check_generated_file_sync.py backend/internal/handler/api front/tools front/app/lib/backend-api/openapi
  git commit -m "feat: ToDo API契約と生成を追加"
  ```

### Task 3: Clean Architecture で ToDo のドメインと更新競合を実装する

**Files:**

- Create: `backend/go.mod`
- Create: `backend/internal/entity/todo.go`
- Create: `backend/internal/entity/todo_test.go`
- Create: `backend/internal/usecase/todo/repository.go`
- Create: `backend/internal/usecase/todo/usecase.go`
- Create: `backend/internal/usecase/todo/usecase_test.go`
- Create: `backend/internal/usecase/todo/errors.go`
- Create: `backend/internal/repository/todo.go`
- Create: `backend/internal/handler/todo.go`
- Create: `backend/db/migrations/20260722000000_create_todos.up.sql`
- Create: `backend/db/migrations/20260722000000_create_todos.down.sql`

**Consumes:** Task 2 の `CreateTodoRequest`、`UpdateTodoRequest`、`Todo`、`Problem`。

**Produces:** `TodoUseCase.Create`、`TodoUseCase.List`、`TodoUseCase.Update`、`TodoUseCase.Delete`。`ErrConflict` は handler で HTTP 409 に変換される。

- [ ] **Step 1: entity の失敗テストを書く**

  `todo_test.go` に以下を追加する。

  ```go
  func TestNewTodoRejectsBlankTitle(t *testing.T) {
      _, err := entity.NewTodo(uuid.New(), "   ", false, 1, nil, now, now)
      require.ErrorIs(t, err, entity.ErrEmptyTodoTitle)
  }
  ```

  Run: `cd backend && go test ./internal/entity -run TestNewTodoRejectsBlankTitle`

  Expected: `entity.NewTodo` が未定義のため FAIL。

- [ ] **Step 2: entity を実装する**

  `Todo` は `ID uuid.UUID`、`Title TodoTitle`、`Completed bool`、`Version int32`、`DueDate *time.Time`、`CreatedAt time.Time`、`UpdatedAt time.Time` を持つ。`NewTodo` は空白のみのタイトルを `ErrEmptyTodoTitle` で拒否する。`ReconstructTodo` は repository 専用の DB 復元用関数として提供する。

- [ ] **Step 3: entity テストの成功を確認する**

  Run: `cd backend && go test ./internal/entity`

  Expected: PASS。

- [ ] **Step 4: usecase の競合テストを書く**

  `usecase_test.go` に、repository stub が `ErrVersionConflict` を返すと `TodoUseCase.Update` が `todo.ErrConflict` を返すテストを書く。

  ```go
  func TestUpdateMapsVersionConflict(t *testing.T) {
      repo := &stubRepository{updateErr: todo.ErrVersionConflict}
      usecase := todo.NewTodoUseCase(repo, fixedClock)
      _, err := usecase.Update(ctx, todo.UpdateInput{ID: id, Version: 3})
      require.ErrorIs(t, err, todo.ErrConflict)
  }
  ```

  Run: `cd backend && go test ./internal/usecase/todo -run TestUpdateMapsVersionConflict`

  Expected: FAIL。

- [ ] **Step 5: repository interface と usecase を実装する**

  `Repository` は `Create(context.Context, *entity.Todo) error`、`List(context.Context) ([]*entity.Todo, error)`、`Update(context.Context, *entity.Todo, int32) error`、`Delete(context.Context, uuid.UUID) error` を定義する。`Update` は `WHERE id = $1 AND version = $2` を使い、影響行が 0 のとき `ErrVersionConflict` を返す。

- [ ] **Step 6: migration と PostgreSQL repository を実装する**

  up migration は `todos` テーブルに `id uuid primary key`、`title text not null`、`completed boolean not null default false`、`version integer not null default 1`、`due_date timestamptz null`、`created_at timestamptz not null`、`updated_at timestamptz not null` を作成する。down migration は `DROP TABLE todos;` のみを行う。

- [ ] **Step 7: handler の HTTP テストを書く**

  `handler/todo_test.go` で、更新 usecase が `todo.ErrConflict` を返す場合に `PATCH /todos/{id}` が `409` と `application/problem+json` を返すことを確認する。

- [ ] **Step 8: handler と Problem Details を実装する**

  handler は生成済み API 型から request を読み、usecase を呼ぶ。`ErrConflict` を `{ "type": "https://example.invalid/problems/version-conflict", "title": "Todo was modified", "status": 409 }` に変換する。空タイトルは 400、存在しない ID は 404、想定外の error は詳細を含めない 500 に変換する。

- [ ] **Step 9: backend テストを確認する**

  Run: `cd backend && go test ./internal/entity ./internal/usecase/todo ./internal/handler`

  Expected: PASS。

- [ ] **Step 10: コミットする**

  ```bash
  git add backend
  git commit -m "feat: ToDoの同期CRUDを追加"
  ```

### Task 4: Go のアーキテクチャ・静的解析・coverage を強制する

**Files:**

- Create: `backend/.golangci.yml`
- Create: `backend/.go-arch-lint.yml`
- Create: `backend/cmd/coverage-gate/main.go`
- Create: `backend/cmd/coverage-gate/main_test.go`
- Create: `backend/scripts/run_coverage_gate.sh`
- Create: `backend/cmd/backend-lint/main.go`
- Create: `backend/internal/lint/analyzers/analyzers.go`
- Create: `docs/adr/0002-clean-architecture-quality-gates.md`

**Consumes:** Task 3 の package 配置。

**Produces:** `go-arch-lint check`、`./scripts/run_coverage_gate.sh`、`go run ./cmd/backend-lint ./...`。

- [ ] **Step 1: architecture violation fixture を書く**

  `backend/internal/lint/analyzers` に、handler package が repository を import する fixture を追加する。

  Run: `cd backend && go-arch-lint check`

  Expected: fixture の違反を含むため FAIL。

- [ ] **Step 2: go-arch-lint 設定を実装する**

  `.go-arch-lint.yml` は `entity` を project dependency 禁止、`usecase` を project dependency 禁止、`repository` を `usecase` と `repository/internal/**` のみ許可、`handler` を `usecase` と `handler/api/**`、`handler/problem/**` のみ許可とする。`.*_test.go` は検査対象外にする。

- [ ] **Step 3: gocognit 設定を実装する**

  `.golangci.yml` を次の値で作る。

  ```yaml
  version: "2"
  linters:
    default: standard
    enable: [gocognit]
    settings:
      gocognit:
        min-complexity: 15
  ```

- [ ] **Step 4: coverage gate の失敗テストを書く**

  `main_test.go` に、`internal/usecase/todo/usecase.go` の未実行ブロックによる scoped coverage `84.9` が threshold `85.0` で exit code 1 になるテストを書く。

  Run: `cd backend && go test ./cmd/coverage-gate -run TestRunFailsBelowThreshold`

  Expected: FAIL。

- [ ] **Step 5: coverage gate を実装する**

  `cmd/coverage-gate/main.go` は Go coverprofile を読み、`cmd/`、`internal/handler/api/`、`internal/repository/internal/db/`、`internal/lint/` を除外して statement coverage を小数第1位に丸める。`-threshold` の既定値は `85.0` とする。

  `run_coverage_gate.sh` は `./scripts/run_go_test.sh -p 1 -coverprofile=<tempfile> ./...` を実行した後、`go run ./cmd/coverage-gate -coverprofile <tempfile> -threshold 85.0` を実行する。

- [ ] **Step 6: custom lint を実装する**

  `backend-lint` は Go analysis multichecker とし、repository 以外の package から `entity.Reconstruct*` を呼ぶコードを失敗させる analyzer を登録する。

- [ ] **Step 7: 全品質ゲートの成功を確認する**

  Run: `cd backend && gofmt -s -l . && golangci-lint run ./cmd/... ./internal/... && go-arch-lint check && go run ./cmd/backend-lint ./... && ./scripts/run_coverage_gate.sh`

  Expected: format 対象の出力が空で、各コマンドが exit code 0。

- [ ] **Step 8: コミットする**

  ```bash
  git add backend/.golangci.yml backend/.go-arch-lint.yml backend/cmd backend/internal/lint backend/scripts docs/adr/0002-clean-architecture-quality-gates.md
  git commit -m "build: Go品質ゲートを追加"
  ```

### Task 5: React の ToDo UI と Frontend 品質ゲートを実装する

**Files:**

- Create: `front/package.json`
- Create: `front/eslint.config.ts`
- Create: `front/app/features/todos/api.ts`
- Create: `front/app/features/todos/TodoList.tsx`
- Create: `front/app/features/todos/TodoForm.tsx`
- Create: `front/app/features/todos/TodoList.test.tsx`
- Create: `front/app/routes/todos.tsx`
- Create: `front/app/routes/todos.test.tsx`
- Create: `front/scripts/coverage_gate.py`
- Create: `front/scripts/run_coverage_gate.sh`

**Consumes:** Task 2 の生成済み TypeScript API 型、Task 3 の `GET/POST/PATCH/DELETE /todos`。

**Produces:** `/todos` で作成、完了切替、競合時の再読み込みをできる UI と `npm run typecheck`、`npm run lint`、`npm run fmt:check`、`npm run coverage:gate`。

- [ ] **Step 1: ToDo 一覧の失敗テストを書く**

  `TodoList.test.tsx` に MSW を使ったテストを書く。

  ```tsx
  it("marks a todo complete", async () => {
    render(<TodoList initialTodos={[todo]} />)
    await userEvent.click(screen.getByRole("checkbox", { name: /buy milk/i }))
    expect(await screen.findByText("Completed")).toBeInTheDocument()
  })
  ```

  Run: `cd front && mise exec node -- npm run test -- TodoList.test.tsx`

  Expected: `TodoList` が未定義のため FAIL。

- [ ] **Step 2: feature を実装する**

  `api.ts` は生成済み型だけを import し、`listTodos`、`createTodo`、`updateTodo`、`deleteTodo` を提供する。`TodoList` は checkbox 操作時に現在の `version` を送る。409 を受けたら「他の更新がありました。最新の内容を読み込みます。」を表示し、一覧を再取得する。`TodoForm` は空白タイトルを送信前に拒否する。

- [ ] **Step 3: route を実装してテストを成功させる**

  `/todos` route は server loader で一覧を取得し、client action で作成・更新・削除を行う。`todos.test.tsx` は作成後に一覧へ表示されること、409 後に再取得されることを確認する。

  Run: `cd front && mise exec node -- npm run test -- TodoList.test.tsx todos.test.tsx`

  Expected: PASS。

- [ ] **Step 4: 静的品質スクリプトを設定する**

  `package.json` に次を追加する。

  ```json
  {
    "scripts": {
      "typecheck": "react-router typegen && tsc",
      "lint": "eslint .",
      "fmt:check": "prettier --check .",
      "test": "vitest run",
      "coverage": "vitest run --coverage",
      "coverage:gate": "bash ./scripts/run_coverage_gate.sh"
    }
  }
  ```

  `coverage_gate.py` は UI component、生成 API、root、route 定義、test / mock を除外した line coverage を `60.0%` 以上にし、変更した `app/routes/**` と `app/features/*/api/server.ts` は `80.0%` 以上にする。

- [ ] **Step 5: frontend の全ゲートを検証する**

  Run: `cd front && mise exec node -- npm run typecheck && mise exec node -- npm run lint && mise exec node -- npm run fmt:check && mise exec node -- npm run coverage:gate`

  Expected: すべて exit code 0。coverage 出力に scoped coverage `60.0%` 以上を含む。

- [ ] **Step 6: コミットする**

  ```bash
  git add front
  git commit -m "feat: ToDoフロントエンドを追加"
  ```

### Task 6: ローカル統合・E2E・CI を完成させる

**Files:**

- Create: `backend/docker-compose.yml`
- Create: `backend/.env.example`
- Create: `front/.env.example`
- Create: `scripts/setup-local-worktree-env-var.sh`
- Create: `scripts/run-e2e-local.sh`
- Create: `front/e2e/todos.e2e.spec.ts`
- Create: `.github/workflows/backend_ci.yml`
- Create: `.github/workflows/front_ci.yml`
- Create: `.github/workflows/e2e_ci.yml`
- Create: `docs/adr/0003-test-strategy-and-quality-gates.md`

**Consumes:** Task 1 から Task 5 の起動、品質ゲート、ToDo フロー。

**Produces:** worktree ごとに分離された PostgreSQL、ローカル E2E、path-filter 付き CI。

- [ ] **Step 1: E2E の失敗する仕様を書く**

  `todos.e2e.spec.ts` に以下を追加する。

  ```ts
  test("creates and completes a todo", async ({ page }) => {
    await page.goto("/todos")
    await page.getByLabel("Title").fill("Buy milk")
    await page.getByRole("button", { name: "Add todo" }).click()
    await page.getByRole("checkbox", { name: "Buy milk" }).check()
    await expect(page.getByText("Completed")).toBeVisible()
  })
  ```

  Run: `cd front && mise exec node -- npx playwright test e2e/todos.e2e.spec.ts`

  Expected: server が未起動のため FAIL。

- [ ] **Step 2: ローカル依存サービスと worktree 分離を実装する**

  `backend/docker-compose.yml` は PostgreSQL 17 を提供する。`setup-local-worktree-env-var.sh` は worktree 絶対パスの安定ハッシュから `POSTGRES_PORT`、`BACKEND_PORT`、`FRONTEND_PORT`、`COMPOSE_PROJECT_NAME` を導出し、`backend/.env` と `front/.env` を作成する。固定 `container_name` は設定しない。

- [ ] **Step 3: E2E 起動スクリプトを実装する**

  `run-e2e-local.sh` は専用の Compose project と開発用から `1000` ずらしたポートを使う。PostgreSQL の health check、migration、backend、frontend の `/health` を待ってから Playwright を実行し、開発用のサービスを停止しない。

- [ ] **Step 4: E2E 成功を確認する**

  Run: `./scripts/run-e2e-local.sh`

  Expected: `todos.e2e.spec.ts` が PASS し、開発用 Compose project が停止されない。

- [ ] **Step 5: CI を実装する**

  `backend_ci.yml` は `backend/**` と `shared/backend_api/**` の変更で、migration lint、`gofmt -s`、`golangci-lint`、`go vet`、`backend-lint`、PostgreSQL を使う coverage、migration の `up -> down -> up` を実行する。

  `front_ci.yml` は `front/**` と `shared/backend_api/**` の変更で、typecheck、lint、format、Vitest、coverage gate、生成 client 同期を実行する。

  `e2e_ci.yml` は `backend/**`、`front/**`、`shared/backend_api/**` の変更で PostgreSQL、backend、frontend を起動して Playwright を実行する。全 workflow は `concurrency.cancel-in-progress: true` を設定する。

- [ ] **Step 6: CI 構文を検証する**

  Run: `ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }' .github/workflows/backend_ci.yml .github/workflows/front_ci.yml .github/workflows/e2e_ci.yml`

  Expected: exit code 0。

- [ ] **Step 7: 最終ローカル検証を実行する**

  Run: `make gen && make check-gen && make check && make e2e`

  Expected: すべて exit code 0。

- [ ] **Step 8: コミットする**

  ```bash
  git add backend/docker-compose.yml backend/.env.example front/.env.example scripts front/e2e .github/workflows docs/adr/0003-test-strategy-and-quality-gates.md
  git commit -m "ci: ToDoリファレンスの検証を追加"
  ```

## 本計画の自己レビュー

- 設計書のモノレポ、契約単一ソース、層の依存方向、楽観ロック、Problem Details、品質ゲート、ローカル worktree 分離、主要 E2E は Task 1〜6 に対応する。
- Transactional Outbox、Worker、DLQ、Runbook、Terraform 配備は独立した後続計画へ明示的に分割しており、本計画の完了条件には含めない。
- `gocognit: 15`、backend `85.0%`、frontend `60.0%`、高リスク frontend `80.0%` は全体制約と実装タスクに明記した。
- 未決事項を示す仮置きの記述や、未定義の後続タスク参照は含めない。
