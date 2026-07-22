# Reference Pattern: 非同期入力ファイルを Object Storage に分離する

- Status: Reference
- Date: 2026-07-22

## Context

upload API と worker processing が別 process・別時刻で動く場合、worker は request body ではなく永続化された raw input を再読込する必要がある。CSV、JSON、画像、archive などの raw file は業務 record そのものではなく、処理入力、再実行、監査、問い合わせ調査のための artifact である。

raw file を relational database の blob に長期保存すると、database storage、backup、replication、restore、vacuum、read replica の費用と時間が file volume に引きずられる。逆に object storage だけに置き、DB metadata と結び付けなければ、認可、検索、状態管理、孤立 object の判定が困難になる。

## Pattern

1. raw file 本体は private object storage に保存する。
2. DB job row には job ID、original file name、byte size、media type、checksum、bucket/key、status、row count、failure reason、timestamps を保存する。
3. object key は system-generated UUID を使い、利用者入力の file name、email、account name、secret を含めない。
4. key prefix は environment と workload boundary を表現する。authorization を prefix 文字列だけに依存させない。
5. checksum は upload 完了確認、重複検知、worker download 後の integrity verification に使う。
6. upload 前に size、許可 media type、extension、archive 展開上限を検証する。初期上限の例は 20 MiB とし、要件に合わせて明示的に決める。
7. large file では API memory に全量を保持せず、presigned upload または streaming upload を検討する。
8. worker は DB row から object reference を取得し、処理開始前に status transition を原子的に claim する。
9. list API は raw content や長寿命 download URL を返さず、必要な metadata だけを返す。
10. download が必要な場合は request 時に再認可し、短寿命の URL または API streaming を使い、監査 event を残す。
11. retention は object storage lifecycle で管理し、DB job history の retention と意図的に調整する。
12. DB と object storage は distributed transaction にしない。partial failure を冪等な retry と reconciliation で回復する。

## Applicability

次の場合に使う。

- raw input を非同期 worker が後から読む。
- 再処理、監査、support investigation のため元入力が必要である。
- file volume が DB backup/restore の主要因になりうる。
- object ごとに retention、encryption、access log を設定したい。

数 KB の payload を単一 transaction で永続化し、短期間しか保持せず、独立 storage の運用コストが上回る場合は DB 保存も比較対象になる。

## Upload and registration flow

推奨 flow は次のいずれかとする。

### API-mediated upload

API が size/type を検査し object storage へ upload、checksum を確定後に job row を作る。小さい file では単純だが、API memory、timeout、network egress を監視する。

### Presigned upload

API が upload session と期限付き URL を発行し、client が直接 upload する。client は完了 API を呼び、server は object metadata/checksum を検証して job を登録する。未完了 upload の cleanup が必要になる。

job 登録と後続通知の整合性が必要なら Transactional Outbox を使う。ただし object upload と DB transaction は原子的にならない。

## Reconciliation

定期 job は少なくとも次を検出する。

- orphan object: grace period を超えて DB row が参照しない object。例として24時間後に削除候補とする。
- missing object: DB row が参照する object が存在しない。未処理 job を failed にし、原因を記録する。
- checksum mismatch: object 内容が登録時と一致しない。処理せず隔離する。
- stale upload session: upload 開始後、完了されない session。期限切れとして cleanup する。

reconciliation 自体を冪等にし、同じ job/object へ複数回 remediation しても二重削除や不正な state transition を起こさない。

## Security

- bucket/container の public access を禁止する。
- encryption at rest と TLS in transit を必須にする。
- application role は必要な prefix/action に最小化する。
- malware scan、content validation、archive bomb 対策を risk に応じて入れる。
- raw file に個人情報や secret が含まれる前提で access log と retention を設計する。
- original file name は表示用 metadata として escape し、path/key に使わない。

## Rejected alternatives

### DB blob へ長期保存

transaction は単純になるが、backup/restore/replication を file growth に結合する。

### broker message に file content を含める

message size limit、retry cost、DLQ exposure、observability の点で不向きである。message には object reference と checksum を載せる。

### process local disk

再起動、水平 scale、別 host worker で参照できず、durability と cleanup が不明確になる。

### 処理完了直後に必ず削除

再処理・監査要件を満たせない。保持期間は data classification と business requirement で決める。

## Consequences

DB の肥大化を切り離し、object lifecycle と access control を使える。一方で storage IAM、cleanup、reconciliation、download authorization、local test double が必要になる。

## Observability and operations

- upload success/failure、bytes、latency、checksum failure を計測する。
- orphan/missing object、reconciliation failure、scan failure を alert 対象にする。
- lifecycle expiration と DB history retention の差を定期監査する。
- incident 時の hold が必要なら lifecycle deletion を停止できる手順を持つ。

## Example assets in this repository

`infra/terraform/main.tf` は private S3 bucket、public access block、AES256 encryption、expiration、incomplete multipart cleanup の構成例を持つ。upload API、job metadata、reconciliation、malware scan は未実装であるため、この文書は Accepted ADR ではなく Reference Pattern である。
