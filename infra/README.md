# Terraform

この構成は raw CSV を private S3 に保存し、Worker 向けの SQS と DLQ を作成する最小の非同期処理基盤です。

```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform fmt -check
terraform validate
terraform plan
```

`plan` と `apply` は AWS 認証情報を必要とします。`apply` 前に意図しない差分がないことを確認してください。
