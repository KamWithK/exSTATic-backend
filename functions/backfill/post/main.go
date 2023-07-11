package main

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	dynamo_types "github.com/KamWithK/exSTATic-backend"
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

func putRequest(pk string, sk string, userMedia interface{}) *dynamodb.WriteRequest {
	tableKey, keyErr := dynamo_types.GetCompositeKey(pk, sk)
	if keyErr != nil {
		return nil
	}

	writeRequest, writeErr := dynamo_types.PutItemRequest(svc, tableKey, userMedia)

	if writeErr != nil {
		return nil
	}

	return writeRequest
}

func HandleRequest(ctx context.Context, history BackfillArgs) (*dynamo_types.BatchwriteArgs, error) {
	username := history.Username

	if len(username) == 0 {
		err := errors.New("Invalid username")
		log.Info().Err(err).Send()

		return nil, err
	}

	writeRequests := []*dynamodb.WriteRequest{}

	for _, userMedia := range history.MediaEntries {
		writeRequest := putRequest(userMedia.Key.MediaType+"#"+username, userMedia.Key.MediaIdentifier, &userMedia)
		if writeRequest != nil {
			writeRequests = append(writeRequests, writeRequest)
		}
	}

	for _, userMedia := range history.MediaStats {
		writeRequest := putRequest(userMedia.Key.MediaType+"#"+username, dynamo_types.ZeroPadInt64(*userMedia.Date)+"#"+userMedia.Key.MediaIdentifier, &userMedia)
		if writeRequest != nil {
			writeRequests = append(writeRequests, writeRequest)
		}
	}

	if len(writeRequests) == 0 {
		err := errors.New("Error no valid data")
		log.Info().Err(err).Send()

		return nil, err
	}

	return &dynamo_types.BatchwriteArgs{
		WriteRequests: writeRequests,
		TableName:     "media",
		MaxBatchSize:  25,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
