package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	dynamo_types "github.com/KamWithK/exSTATic-backend"
)

var sess *session.Session
var svc *dynamodb.DynamoDB

type DateArgs struct {
	Key      dynamo_types.UserMediaKey `json:"key" binding:"required"`
	DateTime int64                     `json:"datetime" binding:"required"`
}

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, dateArgs DateArgs) (*dynamo_types.UserMediaStat, error) {
	var compositeKey = dynamo_types.CompositeKey{
		PK: dateArgs.Key.MediaType + "#" + dateArgs.Key.Username,
		SK: dynamo_types.ZeroPadInt64(dateArgs.DateTime) + "#" + dateArgs.Key.MediaIdentifier,
	}

	tableKey, keyErr := dynamodbattribute.MarshalMap(compositeKey)
	if keyErr != nil {
		return nil, fmt.Errorf("Error marshalling key: %s", keyErr.Error())
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		return nil, fmt.Errorf("Error getting DynamoDB item: %s", getErr.Error())
	}

	if result.Item == nil || len(result.Item) == 0 {
		return nil, fmt.Errorf("Item not found in table")
	}

	mediaStats := dynamo_types.UserMediaStat{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &mediaStats); unmarshalErr != nil {
		return nil, fmt.Errorf("Error unmarshalling item: %s", unmarshalErr.Error())
	}
	mediaStats.Key = dateArgs.Key

	return &mediaStats, nil
}

func main() {
	lambda.Start(HandleRequest)
}
