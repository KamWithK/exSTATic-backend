package main

import (
	"context"

	"github.com/KamWithK/exSTATic-backend/internal/user_media"
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

func HandleRequest(ctx context.Context, key user_media.UserMediaKey) (*user_media.UserMediaEntry, error) {
	return user_media.GetMediaInfo(svc, key)
}

func main() {
	lambda.Start(HandleRequest)
}
