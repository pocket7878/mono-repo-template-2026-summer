# ADR 0001: モノレポと TypeSpec 契約単一ソース

- Status: Accepted
- Date: 2026-07-22

## Context

Frontend、Backend、インフラは同時に変更されることがあり、API 契約が分散すると実装間の差異を見落としやすい。

## Decision

`backend/`、`front/`、`shared/backend_api/`、`infra/terraform/` を一つのリポジトリで管理する。API 契約の唯一のソースは `shared/backend_api` の TypeSpec とし、OpenAPI、Go adapter、TypeScript 型はそこから生成する。

## Consequences

契約変更と実装変更を一つの変更セットとしてレビューできる。一方で、生成手順と同期検証を開発フローに組み込む必要がある。

## Operational Rules

契約を変更した場合は `make gen` で生成物を更新し、`make check-gen` で再生成しても差分がないことを確認する。
