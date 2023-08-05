package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/KamWithK/exSTATic-backend/internal/dynamo_wrapper"
	"github.com/KamWithK/exSTATic-backend/internal/user_media"
)

func HandleRequest(ctx context.Context, history user_media.BackfillArgs) (*dynamo_wrapper.BatchwriteArgs, error) {
	return user_media.PutBackfill(history)
}

func main() {
	lambda.Start(HandleRequest)
}
