package models

import (
	"errors"
	"time"

	"github.com/KamWithK/exSTATic-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog/log"
)

type ProgressStatus struct {
	DateTime int64 `json:"datetime" binding:"required"`
	Pause    bool  `json:"status_change"`
}

type StatusArgs struct {
	Key        UserMediaKey     `json:"key" binding:"required"`
	Stats      MediaStat        `json:"stats" binding:"required"`
	Progress   []ProgressStatus `json:"progress" binding:"required"`
	Timezone   string           `json:"timezone" binding:"required"`
	MaxAFKTime int16            `json:"max_afk_time"`
}

type IntermediateStatItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
	UserMediaStat
}

var startOfTime = time.Unix(0, 0)
var nightEnd, morningStart = 4, 6

func GetStatusUpdate(svc *dynamodb.DynamoDB, dateArgs UserMediaDateKey) (*UserMediaStat, error) {
	tableKey, keyErr := utils.GetCompositeKey(dateArgs.Key.MediaType+"#"+dateArgs.Key.Username, utils.ZeroPadInt64(dateArgs.DateTime)+"#"+dateArgs.Key.MediaIdentifier)
	if keyErr != nil {
		return nil, keyErr
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		log.Error().Str("table", "media").Interface("key", dateArgs.Key).Msg("Dynamodb failed to get item")
		return nil, getErr
	}

	if result.Item == nil || len(result.Item) == 0 {
		log.Info().Str("table", "media").Interface("key", dateArgs.Key).Msg("Item not in table")
		return nil, errors.New("Item not found in table")
	}

	mediaStats := UserMediaStat{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &mediaStats); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "media").Interface("key", dateArgs.Key).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, unmarshalErr
	}
	mediaStats.Key = dateArgs.Key

	return &mediaStats, nil
}

func getDay(svc *dynamodb.DynamoDB, targetDay int64, key UserMediaKey) (map[string]*dynamodb.AttributeValue, *UserMediaStat, error) {
	// Get key which represents this media today
	tableKey, keyErr := utils.GetCompositeKey(key.MediaType+"#"+key.Username, utils.ZeroPadInt64(targetDay)+"#"+key.MediaIdentifier)
	if keyErr != nil {
		return nil, nil, keyErr
	}

	// Get entry from database if it exists
	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		log.Info().Str("table", "media").Interface("key", key).Msg("Item not in table")
		return nil, nil, getErr
	}

	currentStats := UserMediaStat{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &currentStats); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "media").Interface("key", key).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, nil, unmarshalErr
	}

	currentStats.Key = key
	currentStats.Date = &targetDay

	return tableKey, &currentStats, nil
}

func whichDay(svc *dynamodb.DynamoDB, dateTime int64, timezone string, key UserMediaKey) (map[string]*dynamodb.AttributeValue, *UserMediaStat, error) {
	// Given time
	timeNow := time.Unix(dateTime, 0)

	// Location information
	location, locationErr := time.LoadLocation(timezone)
	if locationErr != nil {
		log.Debug().Err(locationErr).Str("timezone", timezone).Send()
		return nil, nil, locationErr
	}
	localTime := timeNow.In(location)

	// Time markers
	morningMarker := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), morningStart, 0, 0, 0, location)
	eveningMarker := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), nightEnd, 0, 0, 0, location)
	yesterday := time.Date(localTime.Year(), localTime.Month(), localTime.Day()-1, 0, 0, 0, 0, time.UTC)
	today := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, time.UTC)

	// Anything before the evening marker is definitely yesterday
	if localTime.Before(eveningMarker) {
		return getDay(svc, yesterday.Unix(), key)
	}

	// Anything after the morning marker is definitely today
	if localTime.After(morningMarker) {
		return getDay(svc, today.Unix(), key)
	}

	// Get yesterdays stats
	yesterdayTableKey, userMediaStats, getDayErr := getDay(svc, yesterday.Unix(), key)
	if getDayErr != nil {
		return nil, nil, getDayErr
	}

	// Continuous immersion with under an hour break constitutes a continuation of yesterday
	// Otherwise immersion occurs today
	yesterdayLastUpdate := time.Unix(userMediaStats.LastUpdate, 0)
	if timeNow.Before(yesterdayLastUpdate.Add(1 * time.Hour)) {
		return yesterdayTableKey, userMediaStats, nil
	} else {
		return getDay(svc, today.Unix(), key)
	}
}

func processProgress(statusArgs *StatusArgs, previousSats *UserMediaStat, morningStars int) {
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

func PutStatusUpdate(svc *dynamodb.DynamoDB, statusArgs StatusArgs) error {
	timeNow := time.Now()
	givenTime := time.Unix(statusArgs.Progress[0].DateTime, 0)

	// Anti-cheat measure
	if timeNow.Sub(givenTime) > 24*time.Hour {
		err := errors.New("First given time is more than 24 hours in the past")
		log.Info().Err(err).Send()
	}

	// Find day
	tableKey, userMediaStats, findDayErr := whichDay(svc, givenTime.Unix(), statusArgs.Timezone, statusArgs.Key)
	if findDayErr != nil {
		return findDayErr
	}

	// Process time data
	processProgress(&statusArgs, userMediaStats, 4)

	// Put item
	_, updateErr := utils.UpdateItem(svc, "media", tableKey, userMediaStats)
	if updateErr != nil {
		return updateErr
	}

	return nil
}
