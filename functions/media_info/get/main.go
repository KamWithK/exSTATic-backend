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

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, key dynamo_types.UserMediaKey) (*dynamo_types.UserSettings, error) {
	tableKey, keyErr := dynamo_types.GetCompositeKey(key.MediaType+"#"+key.Username, key.MediaIdentifier)
	if keyErr != nil {
		return nil, fmt.Errorf("Error getting table key: %s", keyErr.Error())
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		return nil, fmt.Errorf("Error getting DynamoDB item: %s", getErr.Error())
	}

	if result.Item == nil {
		return nil, fmt.Errorf("Item not found in table")
	}

	optionArgs := dynamo_types.UserSettings{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &optionArgs); unmarshalErr != nil {
		return nil, fmt.Errorf("Error unmarshalling item: %s", unmarshalErr.Error())
	}

	return &optionArgs, nil
}

func main() {
	lambda.Start(HandleRequest)
}
