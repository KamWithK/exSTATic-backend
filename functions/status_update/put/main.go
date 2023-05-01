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

var startOfTime = time.Unix(0, 0)
var nightEnd, morningStart = 4, 6

type ProgressStatus struct {
	DateTime int64 `json:"datetime" binding:"required"`
	Pause    bool  `json:"status_change"`
}

type StatusArgs struct {
	Key        dynamo_types.UserMediaKey `json:"key" binding:"required"`
	Stats      dynamo_types.MediaStat    `json:"stats" binding:"required"`
	Progress   []ProgressStatus          `json:"progress" binding:"required"`
	Timezone   string                    `json:"timezone" binding:"required"`
	MaxAFKTime int16                     `json:"max_afk_time"`
}

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func getDay(targetDay int64, key dynamo_types.UserMediaKey) (map[string]*dynamodb.AttributeValue, *dynamo_types.UserMediaStat, error) {
	// Get key which represents this media today
	var compositeKey = dynamo_types.CompositeKey{
		PK: key.MediaType + "#" + key.Username,
		SK: dynamo_types.ZeroPadInt64(targetDay) + "#" + key.MediaIdentifier,
	}
	tableKey, keyErr := dynamodbattribute.MarshalMap(compositeKey)
	if keyErr != nil {
		return nil, nil, fmt.Errorf("Error marshalling key: %s", keyErr.Error())
	}

	// Get entry from database if it exists
	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		return nil, nil, fmt.Errorf("Error getting DynamoDB item: %s", getErr.Error())
	}

	currentStats := dynamo_types.UserMediaStat{}
	if err := dynamodbattribute.UnmarshalMap(result.Item, &currentStats); err != nil {
		return nil, nil, fmt.Errorf("Error unmarshalling item: %s", err.Error())
	}

	currentStats.Key = key
	currentStats.Date = &targetDay

	return tableKey, &currentStats, nil
}

func whichDay(dateTime int64, timezone string, key dynamo_types.UserMediaKey) (map[string]*dynamodb.AttributeValue, *dynamo_types.UserMediaStat, error) {
	// Given time
	timeNow := time.Unix(dateTime, 0)

	// Location information
	location, locationErr := time.LoadLocation(timezone)
	if locationErr != nil {
		return nil, nil, fmt.Errorf("Error loading location: %s", locationErr.Error())
	}
	localTime := timeNow.In(location)

	// Time markers
	morningMarker := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), morningStart, 0, 0, 0, location)
	eveningMarker := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), nightEnd, 0, 0, 0, location)
	yesterday := time.Date(localTime.Year(), localTime.Month(), localTime.Day()-1, 0, 0, 0, 0, time.UTC)
	today := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, time.UTC)

	// Anything before the evening marker is definitely yesterday
	if localTime.Before(eveningMarker) {
		return getDay(yesterday.Unix(), key)
	}

	// Anything after the morning marker is definitely today
	if localTime.After(morningMarker) {
		return getDay(today.Unix(), key)
	}

	// Get yesterdays stats
	yesterdayTableKey, userMediaStats, err := getDay(yesterday.Unix(), key)
	if err != nil {
		return nil, nil, err
	}

	// Continuous immersion with under an hour break constitutes a continuation of yesterday
	// Otherwise immersion occurs today
	yesterdayLastUpdate := time.Unix(userMediaStats.LastUpdate, 0)
	if timeNow.Before(yesterdayLastUpdate.Add(1 * time.Hour)) {
		return yesterdayTableKey, userMediaStats, nil
	} else {
		return getDay(today.Unix(), key)
	}
}

func processProgress(statusArgs *StatusArgs, previousSats *dynamo_types.UserMediaStat, morningStars int) {
	// Set stats reference
	stats := &previousSats.Stats
	lastTime := time.Unix(previousSats.LastUpdate, 0)

	previousSats.Stats.CharsRead += stats.CharsRead
	previousSats.Stats.LinesRead += stats.LinesRead

	// Consolidate the batch of read times together
	for _, progress := range statusArgs.Progress {
		progressTime := time.Unix(progress.DateTime, 0)
		timeDifference := progressTime.Sub(lastTime)

		// Update time read whilst reading and when times are strictly increasing
		if !previousSats.Pause && timeDifference > 0 && timeDifference < time.Duration(statusArgs.MaxAFKTime)*time.Second {
			stats.TimeRead += int64(timeDifference.Seconds())
		}

		// Last update variables pushed forwards
		lastTime = progressTime
		previousSats.LastUpdate = progress.DateTime
		previousSats.Pause = progress.Pause
	}
}

func HandleRequest(ctx context.Context, statusArgs StatusArgs) error {
	timeNow := time.Now()
	givenTime := time.Unix(statusArgs.Progress[0].DateTime, 0)

	// Anti-cheat measure
	if timeNow.Sub(givenTime) > 24*time.Hour {
		return fmt.Errorf("Time error: First given time is more than 24 hours in the past")
	}

	// Find day
	tableKey, userMediaStats, err := whichDay(givenTime.Unix(), statusArgs.Timezone, statusArgs.Key)
	if err != nil {
		return fmt.Errorf("Error extracting the current day: %s", err)
	}

	// Process time data
	processProgress(&statusArgs, userMediaStats, 4)

	// Put item
	updateExpression, expressionAttributeNames, expressionAttributeValues := dynamo_types.CreateUpdateExpressionAttributes(userMediaStats)

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
