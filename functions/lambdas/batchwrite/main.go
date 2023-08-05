package main

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/KamWithK/exSTATic-backend/internal/dynamo_wrapper"
)

var sess *session.Session
var svc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, args *dynamo_wrapper.BatchwriteArgs) (*dynamo_wrapper.BatchwriteArgs, error) {
	nextArgs := dynamo_wrapper.DistributedBatchWrites(svc, args)

	if len(nextArgs.WriteRequests) == 0 {
		log.Info().Msg("Dynamodb batch operations finished")
		return nil, nil
	}

	err := errors.New("unprocessed items error")
	log.Error().Err(err).Msg("")

	return nextArgs, err
}

func main() {
	lambda.Start(HandleRequest)
}
