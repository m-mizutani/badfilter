package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/badfilter/internal"
	"github.com/pkg/errors"
)

var logger = internal.Logger

type snsMessage struct {
	Message          string `json:"Message"`
	MessageID        string `json:"MessageId"`
	Signature        string `json:"Signature"`
	SignatureVersion string `json:"SignatureVersion"`
	SigningCertURL   string `json:"SigningCertURL"`
	Subject          string `json:"Subject"`
	Timestamp        string `json:"Timestamp"`
	TopicArn         string `json:"TopicArn"`
	Type             string `json:"Type"`
	UnsubscribeURL   string `json:"UnsubscribeURL"`
}

func handleRequest(ctx context.Context, event events.SQSEvent) error {
	logger.WithField("event.records.len", len(event.Records)).Debug("Start handler")

	for _, record := range event.Records {
		logger.WithField("body", record.Body).Debug("extracted")

		var snsMsg snsMessage
		if err := json.Unmarshal([]byte(record.Body), &snsMsg); err != nil {
			return errors.Wrapf(err, "Fail to unmarshal SQS event body: %s", record.Body)
		}

		var s3Event events.S3Event
		if err := json.Unmarshal([]byte(snsMsg.Message), &s3Event); err != nil {
			return errors.Wrapf(err, "Fail to unmarshal SNS message: %s", snsMsg.Message)
		}

		for _, s3record := range s3Event.Records {
			logger.WithField("bucket", s3record.S3.Bucket.Name).WithField("key", s3record.S3.Object.Key).Info("Uploaded object")
		}
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
