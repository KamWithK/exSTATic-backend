package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type UserSettings struct {
	Username            string   `json:"username" binding:"required"`
	MediaType           string   `json:"media_type"`
	ShowOnLeaderboard   *bool    `json:"show_on_leaderboard"`
	InterfaceBlurAmount *float32 `json:"interface_blur_amount"`
	MenuBlurAmount      *float32 `json:"menu_blur_amount"`
	MaxAFKTime          *float32 `json:"max_afk_time"`
	MaxBlurTime         *float32 `json:"max_blur_time"`
	MaxLoadLines        *int16   `json:"max_load_lines"`
}

type TableKey struct {
	Username  string `json:"username" binding:"required"`
	MediaType string `json:"media_type"`
}

var sess *session.Session
var svc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, key TableKey) (*UserSettings, error) {
	tableKey, keyErr := dynamodbattribute.MarshalMap(key)

	if keyErr != nil {
		return nil, fmt.Errorf("Error marshalling key: %s", keyErr.Error())
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("settings"),
		Key:       tableKey,
	})

	if getErr != nil {
		return nil, fmt.Errorf("Error getting DynamoDB item: %s", getErr.Error())
	}

	if result.Item == nil {
		return nil, fmt.Errorf("Item not found in table")
	}

	optionArgs := UserSettings{}
	if err := dynamodbattribute.UnmarshalMap(result.Item, &optionArgs); err != nil {
		return nil, fmt.Errorf("Error unmarshalling item: %s", err.Error())
	}

	return &optionArgs, nil
}

func main() {
	lambda.Start(HandleRequest)
}
