package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jackc/pgx/v5/pgxpool"
	sqsinfra "github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/infra/sqs"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/outbox"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	databaseURL := requiredEnvironment("DATABASE_URL")
	queueURL := requiredEnvironment("OUTBOX_QUEUE_URL")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-northeast-1"
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}
	client := awssqs.NewFromConfig(awsConfig, func(options *awssqs.Options) {
		if endpoint := os.Getenv("AWS_ENDPOINT_URL_SQS"); endpoint != "" {
			options.BaseEndpoint = aws.String(endpoint)
		}
	})
	publisher, err := sqsinfra.NewPublisher(client, queueURL)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	worker := outbox.NewWorker(newOutboxRepository(pool), publisher, 3)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := worker.RunOnce(ctx); err != nil {
				log.Printf("outbox worker run failed: %v", err)
			}
		}
	}
}

func requiredEnvironment(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}
