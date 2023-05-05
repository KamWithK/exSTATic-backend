package main

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	dynamo_types "github.com/KamWithK/exSTATic-backend"
)

var sess *session.Session
var svc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, options dynamo_types.UserSettings) error {
	tableKey, keyErr := dynamodbattribute.MarshalMap(options.Key)
	if keyErr != nil {
		log.Error().Err(keyErr).Str("table", "settings").Interface("key", options.Key).Msg("Could not marshal dynamodb key")
		return keyErr
	}

	_, updateErr := dynamo_types.UpdateItem(svc, "settings", tableKey, options)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
