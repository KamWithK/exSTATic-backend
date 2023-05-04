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

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, args *dynamo_types.BatchwriteArgs) (*dynamo_types.BatchwriteArgs, error) {
	nextArgs := dynamo_types.DistributedBatchWrites(svc, args)

	if len(nextArgs.WriteRequests) == 0 {
		return nil, nil
	}

	return nextArgs, fmt.Errorf("Unprocessed items error")
}

func main() {
	lambda.Start(HandleRequest)
}
