package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/KamWithK/exSTATic-backend/internal/backfill"
	"github.com/KamWithK/exSTATic-backend/internal/dynamo_wrapper"
)

func HandleRequest(ctx context.Context, history backfill.BackfillArgs) (*dynamo_wrapper.BatchwriteArgs, error) {
	return backfill.PutBackfill(history)
}

func main() {
	lambda.Start(HandleRequest)
}
