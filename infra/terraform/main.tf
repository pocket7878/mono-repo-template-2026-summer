provider "aws" {
  region = var.aws_region

  default_tags {
    tags = merge(var.tags, {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
    })
  }
}

locals {
  name_prefix = "${var.project}-${var.environment}"
}

resource "aws_sqs_queue" "todo_import_dlq" {
  name                      = "${local.name_prefix}-todo-import-dlq"
  message_retention_seconds = 1209600
}

resource "aws_sqs_queue" "todo_import" {
  name                       = "${local.name_prefix}-todo-import"
  visibility_timeout_seconds = 60
  receive_wait_time_seconds  = 20

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.todo_import_dlq.arn
    maxReceiveCount     = 3
  })
}

resource "aws_s3_bucket" "raw_import" {
  bucket_prefix = "${local.name_prefix}-todo-import-"
}

resource "aws_s3_bucket_public_access_block" "raw_import" {
  bucket                  = aws_s3_bucket.raw_import.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "raw_import" {
  bucket = aws_s3_bucket.raw_import.id

  rule {
    id     = "expire-raw-imports"
    status = "Enabled"

    filter {}

    expiration {
      days = 30
    }
  }
}
