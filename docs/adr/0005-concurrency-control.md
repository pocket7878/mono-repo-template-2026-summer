# ADR 0005: 更新競合は optimistic locking で明示する

- Status: Accepted
- Date: 2026-07-22

## Context

複数の UI 操作、worker、import が同じ集約を更新しうる。last-write-wins は確認済みの値を黙って失わせる。

## Decision

更新可能な集約は version を持ち、更新 SQL は期待 version を条件にする。更新件数 0 は conflict として API の 409 に変換する。複数テーブルの不変条件は同一 transaction に置く。

## Rejected alternatives

- last-write-wins: 上書きを検知できない。
- 長時間の悲観 lock: UI 操作や外部 I/O と相性が悪い。

## Consequences and rules

client は conflict 時に再読込・再編集できる。version を回避する bulk 更新は、同等の競合制御と理由を明示する。
