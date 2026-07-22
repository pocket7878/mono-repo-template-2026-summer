# ADR 0002: 依存方向と品質ゲート

- Status: Accepted
- Date: 2026-07-22

## Context

層分離は規約だけでは劣化する。品質基準も個人判断に委ねると、変更ごとの検証粒度が揺れる。

## Decision

backend の依存は `handler -> usecase <- repository -> entity` とする。handler は repository を直接呼ばない。`go-arch-lint` で依存方向を、custom lint で repository における `entity.Reconstruct*` の利用を検査する。Go は gocognit 15、scoped coverage 85%、frontend は scoped coverage 60% を下限とする。CI は format、lint、生成同期、unit/integration、E2E を変更範囲に応じて実行する。

## Rejected alternatives

- レビューだけで依存方向を守る: 機械的に検出できる逸脱を見落とす。
- 全体 coverage だけを使う: generated code と起動配線が指標を歪める。
- 閾値を持たない lint: 将来の複雑化を止められない。

## Consequences

設計制約を継続的に検証できる反面、テストと lint の保守が必要になる。除外対象や閾値は都合で変えず、変更理由を ADR に残す。

## Operational rules

- architecture/lint の無効化は原則禁止する。
- `Reconstruct*` は永続化済み値の復元専用で、repository の本番コード以外から使わない。
- 新規機能はどのテスト階層で担保するかを PR に記載する。
