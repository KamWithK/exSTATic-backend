package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	dynamo_types "github.com/KamWithK/exSTATic-backend"
	batchwrite "github.com/KamWithK/exSTATic-backend/batchwrite"
)

var sess *session.Session
var svc *dynamodb.DynamoDB

type BackfillArgs struct {
	Username     string                        `json:"username"`
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
		return nil, fmt.Errorf("Error creating put item request: %s", putErr.Error())
	}

	return writeRequest, nil
}

func HandleRequest(ctx context.Context, history BackfillArgs) (*batchwrite.BatchwriteArgs, error) {
	username := history.Username

	writeRequests := []*dynamodb.WriteRequest{}

	for _, userMedia := range history.MediaEntries {
		writeRequest, putErr := putRequest(userMedia.Key.MediaType+"#"+username, userMedia.Key.MediaIdentifier, &userMedia)
		if putErr != nil {
			log.Printf("Error invalid data: %s", putErr)
			continue
		}

		writeRequests = append(writeRequests, writeRequest)
	}

	for _, userMedia := range history.MediaStats {
		writeRequest, putErr := putRequest(userMedia.Key.MediaType+"#"+username, dynamo_types.ZeroPadInt64(*userMedia.Date)+"#"+userMedia.Key.MediaIdentifier, &userMedia)
		if putErr != nil {
			log.Printf("Error invalid data: %s", putErr)
			continue
		}

		writeRequests = append(writeRequests, writeRequest)
	}

	if len(writeRequests) == 0 {
		return nil, fmt.Errorf("Error no valid data")
	}

	return &batchwrite.BatchwriteArgs{
		WriteRequests: writeRequests,
		TableName:     "media",
		MaxBatchSize:  25,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
