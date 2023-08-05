package user_media

import (
	"errors"
	"time"

	"github.com/KamWithK/exSTATic-backend/internal/dynamo_wrapper"
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

func MediaInfoSK(key UserMediaKey) string {
	return key.MediaIdentifier
}

func GetMediaInfo(svc *dynamodb.DynamoDB, key UserMediaKey) (*UserMediaEntry, error) {
	tableKey, keyErr := dynamo_wrapper.GetCompositeKey(UserMediaPK(key), MediaInfoSK(key))
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
		return nil, errors.New("item not found in table")
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
	userMediaEntry.LastUpdate = time.Now().UTC().Unix()

	tableKey, keyErr := dynamo_wrapper.GetCompositeKey(UserMediaPK(userMediaEntry.Key), MediaInfoSK(userMediaEntry.Key))
	if keyErr != nil {
		return keyErr
	}

	_, updateErr := dynamo_wrapper.UpdateItem(svc, "media", tableKey, userMediaEntry)
	if updateErr != nil {
		return updateErr
	}

	return nil
}
