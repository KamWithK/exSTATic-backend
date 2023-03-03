package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type OptionArguments struct {
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
	Username  string
	MediaType string
}

var sess *session.Session
var svc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, options OptionArguments) {
	key := TableKey{
		Username:  options.Username,
		MediaType: options.MediaType,
	}
	tableKey, keyErr := dynamodbattribute.MarshalMap(key)

	if keyErr != nil {
		log.Fatalln("Error marshalling key:", keyErr)
	}

	tableItem, itemErr := dynamodbattribute.MarshalMap(options)

	if itemErr != nil {
		log.Fatalln("Error marshalling item:", itemErr)
	}

	_, updateErr := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String("settings"),
		Key:                       tableKey,
		ExpressionAttributeValues: tableItem,
	})

	if itemErr != nil {
		log.Fatalln("Error updating DynamoDB item:", updateErr)
	}

	fmt.Println(tableKey)
	fmt.Println(tableItem)
}

func main() {
	lambda.Start(HandleRequest)
}
