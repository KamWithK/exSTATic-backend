package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/KamWithK/exSTATic-backend/internal/user_media"
	"github.com/KamWithK/exSTATic-backend/internal/utils"
)

func HandleRequest(ctx context.Context, history user_media.BackfillArgs) (*utils.BatchwriteArgs, error) {
	return user_media.PutBackfill(history)
}

func main() {
	lambda.Start(HandleRequest)
}
