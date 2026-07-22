output "todo_import_queue_url" {
  value       = aws_sqs_queue.todo_import.url
  description = "Queue URL for the outbox publisher and Worker."
}

output "todo_import_dlq_url" {
  value       = aws_sqs_queue.todo_import_dlq.url
  description = "Dead-letter queue URL for failed import events."
}

output "raw_import_bucket_name" {
  value       = aws_s3_bucket.raw_import.bucket
  description = "Private S3 bucket storing raw CSV uploads."
}
