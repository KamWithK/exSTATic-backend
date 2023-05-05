package main

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
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

func HandleRequest(ctx context.Context, key dynamo_types.UserSettingsKey) (*dynamo_types.UserSettings, error) {
	tableKey, keyErr := dynamodbattribute.MarshalMap(key)
	if keyErr != nil {
		log.Error().Err(keyErr).Str("table", "settings").Interface("key", key).Msg("Could not unmarshal dynamodb key")
		return nil, keyErr
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("settings"),
		Key:       tableKey,
	})
	if getErr != nil {
		log.Error().Str("table", "settings").Interface("key", key).Msg("Dynamodb failed to get item")
		return nil, getErr
	}

	if result.Item == nil || len(result.Item) == 0 {
		log.Info().Str("table", "settings").Interface("key", key).Msg("Item not in table")
		return nil, errors.New("Item not found in table")
	}

	optionArgs := dynamo_types.UserSettings{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &optionArgs); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "settings").Interface("key", key).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, unmarshalErr
	}
	optionArgs.Key = key

	return &optionArgs, nil
}

func main() {
	lambda.Start(HandleRequest)
}
