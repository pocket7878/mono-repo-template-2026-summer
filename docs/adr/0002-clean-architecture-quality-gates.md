# ADR 0002: Clean Architecture と品質ゲート

- Status: Accepted
- Date: 2026-07-22

## Context

境界と品質の規約をレビュー時の注意力だけに任せると、層の依存方向と検証水準が変化する。

## Decision

Go の層を `handler -> usecase <- repository -> entity` に固定し、`go-arch-lint` で検証する。`gocognit` の cognitive complexity は 15、backend scoped coverage は 85.0% を合格基準にする。

## Consequences

局所的には実装量とテスト量が増えるが、複雑な関数・境界違反・未検証の業務ロジックをマージ前に検出できる。

## Operational Rules

coverage 除外は entry point、生成 API、生成 DB コード、lint 実装だけに限定する。例外が必要ならこの ADR を更新する。
