package main

import (
	"context"

	"github.com/KamWithK/exSTATic-backend/models"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var sess *session.Session
var svc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, options models.UserSettings) error {
	return models.PutUserSettings(svc, options)
}

func main() {
	lambda.Start(HandleRequest)
}
