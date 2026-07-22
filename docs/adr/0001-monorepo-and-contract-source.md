# ADR 0001: Monorepo と契約単一ソース

- Status: Accepted
- Date: 2026-07-22

## Context

このリファレンスは Go API、React UI、Terraform、API 契約を同時に扱う。契約を各実装に複製すると、レビュー時に仕様差分と生成物の更新漏れを検知できない。

## Decision

`backend/`、`front/`、`infra/terraform/`、`shared/backend_api/` を同じリポジトリに置く。TypeSpec を API 契約の唯一のソースとし、OpenAPI、Go handler、frontend 型は生成物として扱う。契約変更は TypeSpec、生成物、実装、テストを同じ変更セットで更新し、`make check-gen` でドリフトを拒否する。

## Rejected alternatives

- 実装ごとに OpenAPI を手編集する: 変更元が曖昧になる。
- package registry で契約だけを配布する: 独立リリースの運用が必要になり、小さなチームの同時変更を遅くする。

## Consequences

契約レビューが明示的になり、frontend/backend の変更を原子的に確認できる。一方で生成ツールの理解と CI 時間が必要になる。

## Operational rules

- 生成物を直接編集しない。
- API 変更 PR は TypeSpec の差分と互換性を説明する。
- リポジトリ分割が必要になった時点で、契約の versioning 方針を別 ADR で決める。
