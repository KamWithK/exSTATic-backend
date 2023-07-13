package main

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/KamWithK/exSTATic-backend/utils"
)

var sess *session.Session
var svc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, args *utils.BatchwriteArgs) (*utils.BatchwriteArgs, error) {
	nextArgs := utils.DistributedBatchWrites(svc, args)

	if len(nextArgs.WriteRequests) == 0 {
		log.Info().Msg("Dynamodb batch operations finished")
		return nil, nil
	}

	err := errors.New("Unprocessed items error")
	log.Error().Err(err).Msg("")

	return nextArgs, err
}

func main() {
	lambda.Start(HandleRequest)
}
