package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/KamWithK/exSTATic-backend/internal/models"
	"github.com/KamWithK/exSTATic-backend/internal/utils"
)

func HandleRequest(ctx context.Context, history models.BackfillArgs) (*utils.BatchwriteArgs, error) {
	return models.PutBackfill(history)
}

func main() {
	lambda.Start(HandleRequest)
}
