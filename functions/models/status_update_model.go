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
type ProgressPoints []ProgressStatus

type StatusArgs struct {
	Key      UserMediaKey   `json:"key" binding:"required"`
	Stats    MediaStat      `json:"stats" binding:"required"`
	Progress ProgressPoints `json:"progress" binding:"required"`
	Timezone string         `json:"timezone" binding:"required"`
}

type IntermediateStatItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
	UserMediaStat
}

var nightEnd, morningStart = 4, 6

func StatusUpdateSK(dateKey UserMediaDateKey) string {
	return utils.ZeroPadInt64(dateKey.DateTime) + "#" + dateKey.Key.MediaIdentifier
}

func CustomStatusUpdateSK(key UserMediaKey, date int64) string {
	return utils.ZeroPadInt64(date) + "#" + key.MediaIdentifier
}

func GetStatusUpdate(svc *dynamodb.DynamoDB, dateArgs UserMediaDateKey) (*UserMediaStat, error) {
	tableKey, keyErr := utils.GetCompositeKey(UserMediaPK(dateArgs.Key), StatusUpdateSK(dateArgs))
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
		return nil, errors.New("item not found in table")
	}

	mediaStats := UserMediaStat{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &mediaStats); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "media").Interface("key", dateArgs.Key).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, unmarshalErr
	}
	mediaStats.Key = dateArgs.Key

	return &mediaStats, nil
}

func DeleteStatusUpdate(svc *dynamodb.DynamoDB, dateArgs UserMediaDateKey) error {
	tableKey, keyErr := utils.GetCompositeKey(UserMediaPK(dateArgs.Key), StatusUpdateSK(dateArgs))
	if keyErr != nil {
		return keyErr
	}

	_, deleteErr := svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if deleteErr != nil {
		log.Error().Err(deleteErr).Str("table", "media").Interface("key", dateArgs).Msg("Dynamodb failed to delete item")
		return deleteErr
	}

	return nil
}

func getDay(svc *dynamodb.DynamoDB, targetDay int64, key UserMediaKey) (map[string]*dynamodb.AttributeValue, *UserMediaStat, error) {
	// Get key which represents this media today
	tableKey, keyErr := utils.GetCompositeKey(UserMediaPK(key), CustomStatusUpdateSK(key, targetDay))
	if keyErr != nil {
		return nil, nil, keyErr
	}

	// Get entry from database if it exists
	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		log.Info().Str("table", "media").Interface("key", key).Msg("Attempt to get items errored")
		return nil, nil, getErr
	}

	currentStats := UserMediaStat{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &currentStats); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "media").Interface("key", key).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, nil, unmarshalErr
	}

	currentStats.Key = key
	currentStats.Date = &targetDay

	if result.Item == nil {
		return tableKey, &currentStats, utils.ErrEmptyItems
	}

	return tableKey, &currentStats, nil
}

func DayRollback(timeNow time.Time, yesterdayLastUpdate *time.Time) (time.Time, error) {
	// Time markers
	morningMarker := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), morningStart, 0, 0, 0, timeNow.Location())
	eveningMarker := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), nightEnd, 0, 0, 0, timeNow.Location())
	yesterday := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day()-1, 0, 0, 0, 0, time.UTC)
	today := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, time.UTC)

	// Anything before the evening marker is definitely yesterday
	if timeNow.Before(eveningMarker) {
		return yesterday, nil
	}

	// Anything after the morning marker is definitely today
	if timeNow.After(morningMarker) {
		return today, nil
	}

	// Whether to roll back is dependent on the previous days immersion
	// Today will be returned as a reasonable fallback
	if yesterdayLastUpdate == nil {
		return today, errors.New("not enough information, need yesterday")
	}

	// Continuous immersion with under an hour break constitutes a continuation of yesterday
	if timeNow.UTC().Before(yesterdayLastUpdate.Add(time.Hour)) {
		return yesterday, nil
	}

	// Default to immersing today when other cases all fail
	return today, nil
}

func whichDay(svc *dynamodb.DynamoDB, localTime time.Time, key UserMediaKey) (map[string]*dynamodb.AttributeValue, *UserMediaStat, error) {
	// An error indicates a request for more data
	whichDate, timeErr := DayRollback(localTime, nil)
	if timeErr != nil {
		// Use the returned fallback/reference date of today to fetch yesterday
		yesterday := whichDate.AddDate(0, 0, -1)
		yesterdayTableKey, userMediaStats, getDayErr := getDay(svc, yesterday.Unix(), key)

		// If yesterday did exist then adjust based on it
		// Otherwise the default (today) will be maintained
		if getDayErr == nil {
			yesterdayLastUpdate := time.Unix(userMediaStats.LastUpdate, 0)
			whichDate, _ = DayRollback(localTime, &yesterdayLastUpdate)
		} else if !errors.Is(getDayErr, utils.ErrEmptyItems) {
			return nil, nil, getDayErr
		}

		// Avoid extra database retrieval when it's already fetched
		if whichDate.Equal(yesterday) {
			return yesterdayTableKey, userMediaStats, nil
		}
	}

	return getDay(svc, whichDate.Unix(), key)
}

func processProgress(previousSats *UserMediaStat, additiveStats MediaStat, progressPoints ProgressPoints, maxAFKTime int16) {
	// Set stats reference
	stats := &previousSats.Stats
	lastTime := time.Unix(previousSats.LastUpdate, 0)

	stats.CharsRead += additiveStats.CharsRead
	stats.LinesRead += additiveStats.LinesRead

	// Consolidate the batch of read times together
	for _, progress := range progressPoints {
		progressTime := time.Unix(progress.DateTime, 0)
		timeDifference := progressTime.Sub(lastTime)

		// Update time read whilst reading and when times are strictly increasing
		if !previousSats.Pause && timeDifference > 0 && timeDifference < time.Duration(maxAFKTime)*time.Second {
			stats.TimeRead += int64(timeDifference.Seconds())
		}

		// Last update variables pushed forwards
		lastTime = progressTime
		previousSats.LastUpdate = progress.DateTime
		previousSats.Pause = progress.Pause
	}
}

func PutStatusUpdate(svc *dynamodb.DynamoDB, statusArgs StatusArgs, maxAFKTime int16) error {
	// Load times
	timeNow := time.Now().UTC()
	givenTime := time.Unix(statusArgs.Progress[0].DateTime, 0)

	// Anti-cheat measure
	if timeNow.Sub(givenTime) > 24*time.Hour {
		err := errors.New("first given time is more than 24 hours in the past")
		log.Info().Err(err).Send()
	}

	// Location information
	location, locationErr := time.LoadLocation(statusArgs.Timezone)
	if locationErr != nil {
		log.Debug().Err(locationErr).Str("timezone", statusArgs.Timezone).Msg("Invalid timezone specified")
		return locationErr
	}
	localTime := givenTime.In(location)

	// Find day
	tableKey, userMediaStats, findDayErr := whichDay(svc, localTime, statusArgs.Key)
	if findDayErr != nil && !errors.Is(findDayErr, utils.ErrEmptyItems) {
		return findDayErr
	}

	// Process time data
	processProgress(userMediaStats, statusArgs.Stats, statusArgs.Progress, maxAFKTime)

	// Put item
	_, updateErr := utils.UpdateItem(svc, "media", tableKey, userMediaStats)
	if updateErr != nil {
		return updateErr
	}

	return nil
}
