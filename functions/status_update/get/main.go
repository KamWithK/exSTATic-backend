package main

import (
	"context"
	"errors"

	"github.com/KamWithK/exSTATic-backend/utils"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog/log"
)

var sess *session.Session
var svc *dynamodb.DynamoDB

type DateArgs struct {
	Key      utils.UserMediaKey `json:"key" binding:"required"`
	DateTime int64              `json:"datetime" binding:"required"`
}

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, dateArgs DateArgs) (*utils.UserMediaStat, error) {
	tableKey, keyErr := utils.GetCompositeKey(dateArgs.Key.MediaType+"#"+dateArgs.Key.Username, utils.ZeroPadInt64(dateArgs.DateTime)+"#"+dateArgs.Key.MediaIdentifier)
	if keyErr != nil {
		return nil, keyErr
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		log.Error().Str("table", "media").Interface("key", dateArgs.Key).Msg("Dynamodb failed to get item")
		return nil, getErr
	}

	if result.Item == nil || len(result.Item) == 0 {
		log.Info().Str("table", "media").Interface("key", dateArgs.Key).Msg("Item not in table")
		return nil, errors.New("Item not found in table")
	}

	mediaStats := utils.UserMediaStat{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &mediaStats); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "media").Interface("key", dateArgs.Key).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, unmarshalErr
	}
	mediaStats.Key = dateArgs.Key

	return &mediaStats, nil
}

func main() {
	lambda.Start(HandleRequest)
}
