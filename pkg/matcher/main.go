package matcher

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	env "github.com/caarlos0/env/v6"
	"github.com/m-mizutani/badfilter/internal"
	"github.com/m-mizutani/rlogs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger = internal.Logger

type arguments struct {
	RepoS3Region string `env:"REPO_S3_REGION"`
	RepoS3Bucket string `env:"REPO_S3_BUCKET"`
	RepoS3Prefix string `env:"REPO_S3_PREFIX"`
}

// FindEntity picks up entities (ip addr / domain name) from log data
type FindEntity func(rlogs.LogRecord) ([]string, error)

// RunMatcher is entry point of matcher
func RunMatcher(finder FindEntity, reader *rlogs.Reader) {
	lambda.Start(func(ctx context.Context, event events.SQSEvent) error {
		logger.WithField("event", event).Debug("Start Lambda")

		var args arguments
		if err := env.Parse(&args); err != nil {
			return errors.Wrap(err, "Fail to parse environment variable")
		}

		logger.WithFields(logrus.Fields{
			"finder": finder,
			"reader": *reader,
			"args":   args,
		}).Debug("Start RunMatcher")

		bad, err := internal.BuildBadMan(args.RepoS3Region, args.RepoS3Bucket, args.RepoS3Prefix)
		if err != nil {
			return err
		}

		records, err := extractS3Records(event)
		if err != nil {
			return err
		}

		for _, record := range records {
			logger.WithField("s3record", record).Info("Start inspection")
			for q := range reader.Read(&rlogs.AwsS3LogSource{
				Region: record.AWSRegion,
				Bucket: record.S3.Bucket.Name,
				Key:    record.S3.Object.Key,
			}) {
				targets, err := finder(*q.Log)
				if err != nil {
					return err
				}

				for i := range targets {
					entities, err := bad.Lookup(targets[i])
					if err != nil {
						return err
					}

					if len(entities) > 0 {
						logger.WithField("badmen", entities).Warn("Detected")
					}
				}
			}
		}

		return nil
	})
}
