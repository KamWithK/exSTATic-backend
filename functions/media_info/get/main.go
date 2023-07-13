package main

import (
	"context"
	"errors"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog/log"

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

func HandleRequest(ctx context.Context, key utils.UserMediaKey) (*utils.UserMediaEntry, error) {
	tableKey, keyErr := utils.GetCompositeKey(key.MediaType+"#"+key.Username, key.MediaIdentifier)
	if keyErr != nil {
		return nil, keyErr
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		log.Error().Str("table", "media").Interface("key", key).Msg("Dynamodb failed to get item")
		return nil, getErr
	}

	if result.Item == nil {
		log.Info().Str("table", "media").Interface("key", key).Msg("Item not in table")
		return nil, errors.New("Item not found in table")
	}

	userMediaEntryArgs := utils.UserMediaEntry{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &userMediaEntryArgs); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "media").Interface("key", key).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, unmarshalErr
	}

	return &userMediaEntryArgs, nil
}

func main() {
	lambda.Start(HandleRequest)
}
