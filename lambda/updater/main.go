package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/badfilter/internal"
)

var logger = internal.Logger

func handleRequest(ctx context.Context) error {
	logger.Debug("Hello, Hello, Hello")

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
