package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/KamWithK/exSTATic-backend/models"
	"github.com/KamWithK/exSTATic-backend/utils"
)

func HandleRequest(ctx context.Context, history models.BackfillArgs) (*utils.BatchwriteArgs, error) {
	return models.PutBackfill(history)
}

func main() {
	lambda.Start(HandleRequest)
}
