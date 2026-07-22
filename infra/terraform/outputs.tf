output "outbox_queue_url" {
  value       = aws_sqs_queue.outbox.url
  description = "Queue URL for the outbox publisher and Worker."
}

output "outbox_dlq_url" {
  value       = aws_sqs_queue.outbox_dlq.url
  description = "Dead-letter queue URL for failed outbox events."
}

output "raw_import_bucket_name" {
  value       = aws_s3_bucket.raw_import.bucket
  description = "Optional private S3 bucket for the raw-file reference pattern."
}
