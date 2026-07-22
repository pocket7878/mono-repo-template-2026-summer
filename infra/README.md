# Terraform

この構成は Transactional Outbox publisher 向けの SQS と DLQ、監視 alarm を作成する最小の非同期処理基盤です。private S3 bucket は、別途記載した raw-file Reference Pattern を採用する場合の構成例であり、現在の ToDo フローからは使用しません。

```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform fmt -check
terraform validate
terraform plan
```

`plan` と `apply` は AWS 認証情報を必要とします。`apply` 前に意図しない差分がないことを確認してください。

このディレクトリは未適用の新規 reference stack を前提にしています。既存 state へ取り込む場合、queue の resource address や name の変更は replacement になり得るため、先に message を drain し、必要なら `moved` block または `terraform state mv` を用いた移行計画を作成してください。
