package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/badfilter/internal"
)

var logger = internal.Logger

func handleRequest(ctx context.Context, event events.SQSEvent) error {
	logger.WithField("event.records.len", len(event.Records)).Debug("Start handler")

	for _, record := range event.Records {
		logger.WithField("body", record.Body).Debug("extracted")
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
