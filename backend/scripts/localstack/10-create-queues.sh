#!/bin/sh

set -eu

dlq_url="$(awslocal sqs create-queue --queue-name reference-todo-outbox-dlq --query QueueUrl --output text)"
dlq_arn="$(awslocal sqs get-queue-attributes --queue-url "${dlq_url}" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)"
printf '{"RedrivePolicy":"{\"deadLetterTargetArn\":\"%s\",\"maxReceiveCount\":\"3\"}"}' "${dlq_arn}" >/tmp/reference-todo-redrive.json
awslocal sqs create-queue \
  --queue-name reference-todo-outbox \
  --attributes file:///tmp/reference-todo-redrive.json \
  >/dev/null
