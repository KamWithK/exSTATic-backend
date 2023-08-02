package models

import (
	"github.com/KamWithK/exSTATic-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog/log"
)

type UserMediaStateKey struct {
	Username string `json:"username" binding:"required"`
}

type UserMediaState struct {
	Key        UserMediaStateKey `json:"key" binding:"required"`
	LastUpdate int64             `json:"last_update"`
}

func GetUserMediaState(svc *dynamodb.DynamoDB, key UserMediaStateKey) (*UserMediaState, error) {
	tableKey, keyErr := utils.GetCompositeKey(key.Username, "")
	if keyErr != nil {
		return nil, keyErr
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("media"),
		Key:       tableKey,
	})
	if getErr != nil {
		log.Error().Str("table", "media").Str("username", key.Username).Msg("Dynamodb failed to get progress")
		return nil, getErr
	}

	if result.Item == nil || len(result.Item) == 0 {
		return &UserMediaState{
			Key: key,
		}, nil
	}

	var userMediaState UserMediaState
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &userMediaState); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "media").Interface("username", key.Username).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, unmarshalErr
	}

	userMediaState.Key = key

	return &userMediaState, nil
}
