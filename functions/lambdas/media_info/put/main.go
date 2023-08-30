package main

import (
	"context"
	"time"

	"github.com/KamWithK/exSTATic-backend/internal/user_media"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type userMediaEntryArgs struct {
	user_media.UserMediaKey
	user_media.UserMediaEntry
}

var sess *session.Session
var svc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, userMediaEntry userMediaEntryArgs) error {
	return user_media.PutMediaInfo(svc, userMediaEntry.UserMediaKey, userMediaEntry.UserMediaEntry, time.Now().Unix())
}

func main() {
	lambda.Start(HandleRequest)
}
