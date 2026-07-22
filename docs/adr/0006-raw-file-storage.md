# ADR 0006: 非同期入力ファイルは object storage に置く

- Status: Accepted
- Date: 2026-07-22

## Context

upload と worker 処理が分離される場合、worker は元の入力を再読込する。raw file を DB に長期保存すると backup、replication、restore を肥大化させる。

## Decision

このパターンを採用するプロダクトでは、raw file 本体を private object storage に置き、DB には object reference、checksum、content type、size、状態だけを置く。object key はランダム ID を使い、利用者のファイル名を key に含めない。保持期間は lifecycle policy、DB/object の不整合は冪等な reconciliation で扱う。

## Rejected alternatives

- DB blob: 運用コストを DB に持ち込む。
- broker payload: サイズ制限と再送追跡に不向き。

## Consequences and rules

bucket encryption、最小 IAM、lifecycle、孤立 object の cleanup が必要になる。list API は raw content や無制限の download URL を返さない。
