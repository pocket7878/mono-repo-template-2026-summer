package sqs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/outbox"
)

type sender interface {
	SendMessage(context.Context, *awssqs.SendMessageInput, ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error)
}

type Publisher struct {
	client   sender
	queueURL string
}

func NewPublisher(client sender, queueURL string) (*Publisher, error) {
	if client == nil {
		return nil, errors.New("SQS client is required")
	}
	if queueURL == "" {
		return nil, errors.New("SQS queue URL is required")
	}
	return &Publisher{client: client, queueURL: queueURL}, nil
}

func (publisher *Publisher) Publish(ctx context.Context, event outbox.Event) error {
	body, err := json.Marshal(struct {
		IdempotencyKey string          `json:"idempotencyKey"`
		EventType      string          `json:"eventType"`
		Payload        json.RawMessage `json:"payload"`
		Attempt        int             `json:"attempt"`
	}{
		IdempotencyKey: event.ID.String(),
		EventType:      event.Type,
		Payload:        json.RawMessage(event.Payload),
		Attempt:        event.Attempt,
	})
	if err != nil {
		return fmt.Errorf("marshal outbox event: %w", err)
	}
	messageBody := string(body)
	_, err = publisher.client.SendMessage(ctx, &awssqs.SendMessageInput{
		QueueUrl:    &publisher.queueURL,
		MessageBody: &messageBody,
	})
	if err != nil {
		return fmt.Errorf("send outbox event to SQS: %w", err)
	}
	return nil
}
