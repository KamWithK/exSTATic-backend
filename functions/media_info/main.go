package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

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
	var timeNow = time.Now().Unix()
	userMediaEntry.LastUpdate = &timeNow

	var compositeKey = dynamo_types.CompositeKey{
		PK: userMediaEntry.Key.MediaType + "#" + userMediaEntry.Key.Username,
		SK: userMediaEntry.Key.MediaIdentifier,
	}

	tableKey, keyErr := dynamodbattribute.MarshalMap(compositeKey)

	if keyErr != nil {
		return fmt.Errorf("Error marshalling key: %s", keyErr.Error())
	}

	updateExpression, expressionAttributeNames, expressionAttributeValues := dynamo_types.CreateUpdateExpressionAttributes(userMediaEntry)

	_, updateErr := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String("media"),
		Key:                       tableKey,
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	})

	if updateErr != nil {
		return fmt.Errorf("Error updating DynamoDB item: %s", updateErr.Error())
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
