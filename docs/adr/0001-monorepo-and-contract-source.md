# ADR 0001: Monorepo と TypeSpec 契約単一ソースの採用

- Status: Accepted
- Date: 2026-07-22

## Context

このリファレンスは、Go API/worker、React frontend、Terraform、API 契約を一つの変更単位として扱う。frontend と backend は同じ HTTP 契約へ依存し、契約変更に伴って server interface、client type、実装、テストが同時に変わる。

契約を各コンポーネントへ複製すると、同じフィールドや status code に複数の正本が生まれる。人間のレビューでも AI による変更でも、どれを先に変更すべきか曖昧になり、生成物の手修正、互換性の見落とし、片側だけの更新が起きやすい。

## Decision

1. `backend/`、`front/`、`infra/terraform/`、`shared/backend_api/` を同じ Git repository で管理する。
2. API 契約の唯一のソースを `shared/backend_api/main.tsp` とする。
3. TypeSpec から OpenAPI を生成し、OpenAPI から Go server interface と TypeScript type を生成する。
4. `shared/backend_api/openapi/openapi.yaml`、`backend/internal/handler/api/api.gen.go`、`front/app/lib/backend-api/openapi/types.ts` はレビュー可能な生成物として Git 管理するが、直接編集しない。
5. API 変更は TypeSpec、生成物、server/client 実装、テストを同じ pull request に含める。
6. `make gen` を正規の生成入口、`make check-gen` を生成同期の検証入口とする。
7. infrastructure は同じ repository に置くが、application の import dependency にはしない。共有するのは設定上の契約であり、source-level dependency ではない。

## Applicability

次を満たす場合に適用する。

- 複数コンポーネントが同じ API 契約に依存する。
- 契約と実装が同じチーム、または密に連携する少数チームで変更される。
- 契約変更を一つのレビュー単位で確認する価値が、repository 分割による独立性より大きい。

独立した release cadence、権限境界、巨大な build graph が必要になった場合は、契約 package の versioning と repository 分割を再検討する。

## Rejected alternatives

### frontend と backend が別々に型を定義する

短期的には生成手順が不要だが、optional/required、format、status code が別々に変化する。コンパイルは各 repository 内で成功しても、統合時に壊れるため採用しない。

### OpenAPI YAML を直接正本にする

可能だが、共通 model、operation、error response の再利用性と記述支援を重視して TypeSpec を採用する。OpenAPI は外部ツールとの交換形式として生成する。

### 契約を package registry で配布する

独立 versioning が必要な組織では有効だが、この規模では publish 順序、互換 version、更新 bot の運用が増える。分割条件が成立するまで採用しない。

## Consequences

### Positive

- 契約変更の起点が一箇所になり、レビューで意図を追跡できる。
- server/client の型ずれを生成と CI で早期検出できる。
- application と infrastructure の対応を同じ履歴で参照できる。
- 新しい利用者や AI が、実装から契約を推測せずに済む。

### Negative

- 生成ツールと生成物の更新を開発者全員が理解する必要がある。
- 複数言語の dependency setup が必要になる。
- path filter を設計しないと、無関係な変更でも全 CI が走る。
- breaking change を同一 repository で変更できることと、安全に deploy できることは別問題である。

## Implementation

- 契約: `shared/backend_api/main.tsp`
- TypeSpec 設定: `shared/backend_api/tspconfig.yaml`
- OpenAPI: `shared/backend_api/openapi/openapi.yaml`
- Go generation: `backend/internal/handler/api/generate.go` と `cfg.yaml`
- frontend generation: `front/tools/gen-api-client.sh`
- orchestration: root `Makefile` の `gen` / `check-gen`
- CI path scopes: `.github/workflows/backend-ci.yml`、`front-ci.yml`、`e2e-ci.yml`

`make check-gen` はローカルの `make check` と backend CI の `api-contract-sync` job から実行される。

## Verification

```bash
make gen
make check-gen
```

生成後に tracked output が変わらないことを合格条件とする。契約変更時は backend と frontend の型検査・テストも同時に通す。

## Operational rules

- 生成物へ `DO NOT EDIT` がなくても直接編集しない。
- breaking change は field 削除だけでなく、required 化、format 変更、status code 変更も対象とする。
- deploy 順序が必要な変更は expand/contract で段階移行する。
- README は手順への入口とし、契約仕様を複製しない。

## Revisit conditions

- component ごとに独立した権限・release cadence が必要になったとき。
- repository の build/test 時間が path filter では管理できなくなったとき。
- 外部利用者向けに複数 API version の並行提供が必要になったとき。
