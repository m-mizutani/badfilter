package matcher

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
)

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

func extractS3Records(event events.SQSEvent) ([]events.S3EventRecord, error) {
	var resp []events.S3EventRecord

	for _, record := range event.Records {
		logger.WithField("body", record.Body).Debug("extracted")

		var snsMsg snsMessage
		if err := json.Unmarshal([]byte(record.Body), &snsMsg); err != nil {
			return nil, errors.Wrapf(err, "Fail to unmarshal SQS event body: %s", record.Body)
		}

		var s3Event events.S3Event
		if err := json.Unmarshal([]byte(snsMsg.Message), &s3Event); err != nil {
			return nil, errors.Wrapf(err, "Fail to unmarshal SNS message: %s", snsMsg.Message)
		}

		resp = append(resp, s3Event.Records...)
	}

	return resp, nil
}
