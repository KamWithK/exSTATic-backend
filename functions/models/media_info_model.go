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

type IntermediateEntryItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
	UserMediaEntry
}

func GetMediaInfo(svc *dynamodb.DynamoDB, key UserMediaKey) (*UserMediaEntry, error) {
	tableKey, keyErr := utils.GetCompositeKey(key.MediaType+"#"+key.Username, key.MediaIdentifier)
	if keyErr != nil {
		return nil, keyErr
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		log.Error().Str("table", "media").Interface("key", key).Msg("Dynamodb failed to get item")
		return nil, getErr
	}

	if result.Item == nil {
		log.Info().Str("table", "media").Interface("key", key).Msg("Item not in table")
		return nil, errors.New("Item not found in table")
	}

	userMediaEntryArgs := UserMediaEntry{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &userMediaEntryArgs); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "media").Interface("key", key).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, unmarshalErr
	}
	userMediaEntryArgs.Key = key

	return &userMediaEntryArgs, nil
}

func PutMediaInfo(svc *dynamodb.DynamoDB, userMediaEntry UserMediaEntry) error {
	userMediaEntry.LastUpdate = time.Now().Unix()

	tableKey, keyErr := utils.GetCompositeKey(userMediaEntry.Key.MediaType+"#"+userMediaEntry.Key.Username, userMediaEntry.Key.MediaIdentifier)
	if keyErr != nil {
		return keyErr
	}

	_, updateErr := utils.UpdateItem(svc, "media", tableKey, userMediaEntry)
	if updateErr != nil {
		return updateErr
	}

	return nil
}
