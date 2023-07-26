package models

import (
	"errors"

	"github.com/KamWithK/exSTATic-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog/log"
)

type UserSettingsKey struct {
	Username  string `json:"username" binding:"required"`
	MediaType string `json:"media_type"`
}

type UserSettings struct {
	Key                 UserSettingsKey `json:"key" binding:"required"`
	ShowOnLeaderboard   *bool           `json:"show_on_leaderboard"`
	InterfaceBlurAmount *float32        `json:"interface_blur_amount"`
	MenuBlurAmount      *float32        `json:"menu_blur_amount"`
	MaxAFKTime          *int16          `json:"max_afk_time"`
	MaxBlurTime         *int16          `json:"max_blur_time"`
	MaxLoadLines        *int16          `json:"max_load_lines"`
}

func GetUserSettings(svc *dynamodb.DynamoDB, key UserSettingsKey) (*UserSettings, error) {
	tableKey, keyErr := dynamodbattribute.MarshalMap(key)
	if keyErr != nil {
		log.Error().Err(keyErr).Str("table", "settings").Interface("key", key).Msg("Could not unmarshal dynamodb key")
		return nil, keyErr
	}

	result, getErr := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("settings"),
		Key:       tableKey,
	})
	if getErr != nil {
		log.Error().Str("table", "settings").Interface("key", key).Msg("Dynamodb failed to get item")
		return nil, getErr
	}

	if result.Item == nil || len(result.Item) == 0 {
		log.Info().Str("table", "settings").Interface("key", key).Msg("Item not in table")
		return nil, errors.New("item not found in table")
	}

	optionArgs := UserSettings{}
	if unmarshalErr := dynamodbattribute.UnmarshalMap(result.Item, &optionArgs); unmarshalErr != nil {
		log.Error().Err(unmarshalErr).Str("table", "settings").Interface("key", key).Interface("item", result.Item).Msg("Could not unmarshal dynamodb item")
		return nil, unmarshalErr
	}
	optionArgs.Key = key

	return &optionArgs, nil
}

func PutUserSettings(svc *dynamodb.DynamoDB, options UserSettings) error {
	tableKey, keyErr := dynamodbattribute.MarshalMap(options.Key)
	if keyErr != nil {
		log.Error().Err(keyErr).Str("table", "settings").Interface("key", options.Key).Msg("Could not marshal dynamodb key")
		return keyErr
	}

	_, updateErr := utils.UpdateItem(svc, "settings", tableKey, options)
	if updateErr != nil {
		return updateErr
	}

	return nil
}
