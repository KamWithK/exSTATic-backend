package main

import (
	"context"
	"fmt"
	"strconv"
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

type StatusArgs struct {
	Key          dynamo_types.UserMediaKey `json:"key" binding:"required"`
	Stats        dynamo_types.MediaStat    `json:"stats" binding:"required"`
	DateTime     int64                     `json:"datetime" binding:"required"`
	StatusChange bool                      `json:"status_change" binding:"required"`
	Timezone     string                    `json:"timezone" binding:"required"`
	MaxAFKTime   int16                     `json:"max_afk_time"`
}

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, statusArgs StatusArgs) error {
	timeNow := time.Unix(statusArgs.DateTime, 0)

	location, locationErr := time.LoadLocation(statusArgs.Timezone)
	if locationErr != nil {
		return fmt.Errorf("Error loading location: %s", locationErr.Error())
	}
	localTimeNow := timeNow.In(location)

	// Get 4:00 am local time
	morningMarker := time.Date(localTimeNow.Year(), localTimeNow.Month(), localTimeNow.Day(), 4, 0, 0, 0, location)

	// Use the Unix timestamp of 4 am to mark a day
	var targetDay int64

	// If the current local time is before 4 am then set the current day to yesterday, else to today
	if localTimeNow.Before(morningMarker) {
		targetDay = morningMarker.AddDate(0, 0, -1).Unix()
	} else {
		targetDay = morningMarker.Unix()
	}

	// Get key which represents this media today
	var compositeKey = dynamo_types.CompositeKey{
		PK: statusArgs.Key.MediaType + "#" + statusArgs.Key.Username,
		SK: strconv.FormatInt(targetDay, 10) + "#" + statusArgs.Key.MediaIdentifier,
	}

	tableKey, keyErr := dynamodbattribute.MarshalMap(compositeKey)

	if keyErr != nil {
		return fmt.Errorf("Error marshalling key: %s", keyErr.Error())
	}

	// Get entry from database if it exists
	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})

	if getErr != nil {
		return fmt.Errorf("Error getting DynamoDB item: %s", getErr.Error())
	}

	currentStats := dynamo_types.UserMediaStat{}
	if err := dynamodbattribute.UnmarshalMap(result.Item, &currentStats); err != nil {
		return fmt.Errorf("Error unmarshalling item: %s", err.Error())
	}

	previousUpdate := time.Unix(currentStats.LastUpdate, 0)
	timeDifference := timeNow.Sub(previousUpdate)

	currentStats.Key = statusArgs.Key
	currentStats.Date = &targetDay
	currentStats.LastUpdate = timeNow.Unix()

	currentStats.Stats.CharsRead += statusArgs.Stats.CharsRead
	currentStats.Stats.LinesRead += statusArgs.Stats.LinesRead

	// If the last update is over the threshold update the stats
	// Otherwise set the stats to empty and the current time as last updated
	if timeDifference < time.Duration(statusArgs.MaxAFKTime)*time.Second {
		currentStats.Stats.TimeRead += int64(timeDifference.Seconds())
	}

	// Put item
	updateExpression, expressionAttributeNames, expressionAttributeValues := dynamo_types.CreateUpdateExpressionAttributes(currentStats)

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
