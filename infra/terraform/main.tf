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

resource "aws_sqs_queue" "outbox_dlq" {
  name                      = "${local.name_prefix}-outbox-dlq"
  message_retention_seconds = 1209600
}

resource "aws_sqs_queue" "outbox" {
  name                       = "${local.name_prefix}-outbox"
  visibility_timeout_seconds = 60
  receive_wait_time_seconds  = 20

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.outbox_dlq.arn
    maxReceiveCount     = 3
  })
}

resource "aws_cloudwatch_metric_alarm" "outbox_dlq_visible_messages" {
  alarm_name          = "${local.name_prefix}-outbox-dlq-visible-messages"
  alarm_description   = "Messages in the DLQ require investigation using docs/runbooks/sqs-dlq-replay.md."
  namespace           = "AWS/SQS"
  metric_name         = "ApproximateNumberOfMessagesVisible"
  statistic           = "Maximum"
  period              = 60
  evaluation_periods  = 1
  threshold           = 0
  comparison_operator = "GreaterThanThreshold"
  treat_missing_data  = "notBreaching"

  dimensions = {
    QueueName = aws_sqs_queue.outbox_dlq.name
  }
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

resource "aws_s3_bucket_server_side_encryption_configuration" "raw_import" {
  bucket = aws_s3_bucket.raw_import.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
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

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}
