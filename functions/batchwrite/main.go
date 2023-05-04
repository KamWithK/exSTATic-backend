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

type BatchwriteArgs struct {
	TableName     string                   `json:"table_name"`
	WriteRequests []*dynamodb.WriteRequest `json:"write_requests"`
	MaxBatchSize  int                      `json:"max_batch_size" default:"25"`
}

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, args BatchwriteArgs) ([]*dynamodb.WriteRequest, error) {
	unprocessedWrites := dynamo_types.DistributedBatchWrites(svc, args.TableName, args.WriteRequests, args.MaxBatchSize)

	if len(unprocessedWrites) == 0 {
		return nil, nil
	}

	return unprocessedWrites, fmt.Errorf("Unprocessed items error")
}

func main() {
	lambda.Start(HandleRequest)
}
