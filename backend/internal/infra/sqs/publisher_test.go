package sqs_test

import (
	"context"
	"errors"
	"testing"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	sqsinfra "github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/infra/sqs"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/outbox"
	"github.com/stretchr/testify/require"
)

type fakeSender struct {
	input *awssqs.SendMessageInput
	err   error
}

func (sender *fakeSender) SendMessage(_ context.Context, input *awssqs.SendMessageInput, _ ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error) {
	sender.input = input
	return &awssqs.SendMessageOutput{}, sender.err
}

func TestPublisherSendsStableEnvelope(t *testing.T) {
	sender := &fakeSender{}
	publisher, err := sqsinfra.NewPublisher(sender, "http://localhost/queue")
	require.NoError(t, err)
	eventID := uuid.MustParse("41bb7dd8-20f5-4bd7-93af-2b7163bb7a35")

	err = publisher.Publish(t.Context(), outbox.Event{ID: eventID, Type: "todo.changed", Payload: []byte(`{"todoId":"todo-1"}`), Attempt: 2})

	require.NoError(t, err)
	require.Equal(t, "http://localhost/queue", *sender.input.QueueUrl)
	require.JSONEq(t, `{"idempotencyKey":"41bb7dd8-20f5-4bd7-93af-2b7163bb7a35","eventType":"todo.changed","payload":{"todoId":"todo-1"},"attempt":2}`, *sender.input.MessageBody)
}

func TestPublisherPropagatesSendFailure(t *testing.T) {
	sender := &fakeSender{err: errors.New("unavailable")}
	publisher, err := sqsinfra.NewPublisher(sender, "http://localhost/queue")
	require.NoError(t, err)

	err = publisher.Publish(t.Context(), outbox.Event{ID: uuid.New(), Type: "todo.changed", Payload: []byte(`{}`)})

	require.ErrorContains(t, err, "send outbox event to SQS")
}

func TestNewPublisherRejectsMissingConfiguration(t *testing.T) {
	_, err := sqsinfra.NewPublisher(nil, "")
	require.Error(t, err)
}
