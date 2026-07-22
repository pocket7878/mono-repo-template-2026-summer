# ADR 0004: API エラーを Problem Details に統一する

- Status: Accepted
- Date: 2026-07-22

## Context

エラー形式が handler ごとに異なると、UI、契約、テストが個別実装に結び付く。

## Decision

エラーは `application/problem+json` で返し、`type`、`title`、`detail`、`status` を契約に含める。500 の detail は固定文言にし、例外・SQL・スタックトレースを公開しない。独自の `error` ラッパーや、用途のない error code は追加しない。

## Rejected alternatives

- handler ごとの JSON: client の分岐が増える。
- 内部エラーを detail に出す: 情報漏えいになる。

## Consequences and rules

client は status と Problem Details を一貫して扱える。メタデータが必要なら互換性と認可を検討する ADR を追加し、契約・生成物・実装を同時に更新する。
