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

func HandleRequest(ctx context.Context, dateArgs models.UserMediaDateKey) (*models.UserMediaStat, error) {
	return models.GetStatusUpdate(svc, dateArgs)
}

func main() {
	lambda.Start(HandleRequest)
}
