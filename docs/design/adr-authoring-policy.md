# Architecture Decision Documentation Policy

## Goal

このリポジトリの設計判断を、人間と AI の双方が別プロジェクトで再利用できる情報密度で保存する。プロダクト固有語は一般化するが、判断理由、境界条件、却下案、実装制約、検証方法は要約しない。

## Document types

### Accepted ADR

このサンプル実装が実際に採用し、コードまたは CI で検証している判断を記録する。各文書は Context、Decision、Applicability、Rejected alternatives、Consequences、Implementation、Verification、Operational rules、Revisit conditions を持つ。

### Reference Pattern

多くのソフトウェアで再利用できるが、このサンプルでは end-to-end 実装されていない判断を記録する。採用済みと誤読されないよう Status を `Reference` とし、適用条件・非適用条件・必要な実装・セキュリティ・運用を完全に記載する。

## Rules

- 固有名詞を削除するために技術的詳細を削除しない。
- `Problem Details`、at-least-once、Transactional Outbox、optimistic locking など、一般化済みのパターン名は維持する。
- 閾値、HTTP status、message field、SQL 制約、コマンド、対象パスを具体的に記載する。
- 実装済みという主張にはコード、テスト、CI の参照を付ける。
- 選択肢を採用しなかった理由と、将来再検討する条件を記録する。
- 要約版を正本にしない。README は正本への索引だけを持つ。
