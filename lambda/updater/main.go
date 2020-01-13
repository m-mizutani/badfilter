package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	env "github.com/caarlos0/env/v6"
	"github.com/m-mizutani/badfilter/internal"
	"github.com/pkg/errors"
)

var logger = internal.Logger

type arguments struct {
	RepoS3Region string `env:"REPO_S3_REGION"`
	RepoS3Bucket string `env:"REPO_S3_BUCKET"`
	RepoS3Prefix string `env:"REPO_S3_PREFIX"`
}

func handleRequest(ctx context.Context) error {
	var args arguments
	if err := env.Parse(&args); err != nil {
		return errors.Wrap(err, "Fail to parse environment variable")
	}
	logger.WithField("args", args).Info("Start Lambda")

	if err := internal.UpdateBadMan(args.RepoS3Region, args.RepoS3Bucket, args.RepoS3Prefix); err != nil {
		return err
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
