package main

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

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

func HandleRequest(ctx context.Context, userMediaEntry dynamo_types.UserMediaEntry) error {
	userMediaEntry.LastUpdate = time.Now().Unix()

	tableKey, keyErr := dynamo_types.GetCompositeKey(userMediaEntry.Key.MediaType+"#"+userMediaEntry.Key.Username, userMediaEntry.Key.MediaIdentifier)
	if keyErr != nil {
		return keyErr
	}

	_, updateErr := dynamo_types.UpdateItem(svc, "media", tableKey, userMediaEntry)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
