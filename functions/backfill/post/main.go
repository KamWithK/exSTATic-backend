package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	dynamo_types "github.com/KamWithK/exSTATic-backend"
)

var sess *session.Session
var svc *dynamodb.DynamoDB

type BackfillArgs struct {
	Username     string                        `json:"username"`
	GetAfter     int64                         `json:"get_after"`
	MediaEntries []dynamo_types.UserMediaEntry `json:"media_entries"`
	MediaStats   []dynamo_types.UserMediaStat  `json:"media_stats"`
}

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func putRequest(pk string, sk string, userMedia interface{}) (*dynamodb.WriteRequest, error) {
	tableKey, keyErr := dynamo_types.GetCompositeKey(pk, sk)
	if keyErr != nil {
		return nil, fmt.Errorf("Error getting table key: %s", keyErr.Error())
	}

	writeRequest, putErr := dynamo_types.PutItemRequest(svc, tableKey, userMedia)
	if putErr != nil {
		return nil, fmt.Errorf("Error putting DynamoDB item: %s", putErr.Error())
	}

	return writeRequest, nil
}

func HandleRequest(ctx context.Context, history BackfillArgs) ([]*dynamodb.WriteRequest, []error) {
	username := history.Username

	writeRequests := []*dynamodb.WriteRequest{}
	errors := []error{}

	for _, userMedia := range history.MediaEntries {
		writeRequest, putErr := putRequest(userMedia.Key.MediaType+"#"+username, userMedia.Key.MediaIdentifier, &userMedia)
		if putErr != nil {
			errors = append(errors, putErr)
			continue
		}
		writeRequests = append(writeRequests, writeRequest)
	}

	for _, userMedia := range history.MediaStats {
		writeRequest, putErr := putRequest(userMedia.Key.MediaType+"#"+username, dynamo_types.ZeroPadInt64(*userMedia.Date)+"#"+userMedia.Key.MediaIdentifier, &userMedia)
		if putErr != nil {
			errors = append(errors, putErr)
			continue
		}
		writeRequests = append(writeRequests, writeRequest)
	}

	unprocessedWrites, batchErrs := dynamo_types.BatchWriteItems(svc, "media", writeRequests, 25)

	return unprocessedWrites, append(errors, batchErrs...)
}

func main() {
	lambda.Start(HandleRequest)
}
