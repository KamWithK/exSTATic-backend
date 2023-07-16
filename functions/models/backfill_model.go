package models

import (
	"errors"

	"github.com/KamWithK/exSTATic-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog/log"
)

type BackfillArgs struct {
	Username     string           `json:"username"`
	MediaEntries []UserMediaEntry `json:"media_entries"`
	MediaStats   []UserMediaStat  `json:"media_stats"`
}

func GetBackfill(svc *dynamodb.DynamoDB, UserMediaDateKey UserMediaDateKey) ([]UserMediaStat, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("media"),
		KeyConditionExpression: aws.String("pk = :pk AND last_update >= :lastUpdate"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(UserMediaDateKey.Key.MediaType + "#" + UserMediaDateKey.Key.Username),
			},
			":lastUpdate": {
				N: aws.String(utils.ZeroPadInt64(UserMediaDateKey.DateTime)),
			},
		},
		IndexName: aws.String("lastUpdatedIndex"),
	}

	result, queryErr := svc.Query(queryInput)
	if queryErr != nil {
		log.Info().Err(queryErr).Send()
		return nil, queryErr
	}

	userMediaStats := []UserMediaStat{}
	unmarshalErr := dynamodbattribute.UnmarshalListOfMaps(result.Items, &userMediaStats)
	if unmarshalErr != nil {
		log.Info().Err(unmarshalErr).Send()
		return nil, unmarshalErr
	}

	return userMediaStats, nil
}

func PutBackfill(history BackfillArgs) (*utils.BatchwriteArgs, error) {
	username := history.Username

	if len(username) == 0 {
		err := errors.New("Invalid username")
		log.Info().Err(err).Send()

		return nil, err
	}

	writeRequests := []*dynamodb.WriteRequest{}

	for _, userMedia := range history.MediaEntries {
		writeRequest := utils.PutRawRequest(userMedia.Key.MediaType+"#"+username, userMedia.Key.MediaIdentifier, &userMedia)
		if userMedia.Key.Username != username {
			err := errors.New("Username mismatch")
			log.Info().Err(err).Send()
		} else if writeRequest != nil {
			writeRequests = append(writeRequests, writeRequest)
		}
	}

	for _, userMedia := range history.MediaStats {
		writeRequest := utils.PutRawRequest(userMedia.Key.MediaType+"#"+username, utils.ZeroPadInt64(*userMedia.Date)+"#"+userMedia.Key.MediaIdentifier, &userMedia)
		if userMedia.Key.Username != username {
			err := errors.New("Username mismatch")
			log.Info().Err(err).Send()
		} else if writeRequest != nil {
			writeRequests = append(writeRequests, writeRequest)
		}
	}

	if len(writeRequests) == 0 {
		err := errors.New("Error no valid data")
		log.Info().Err(err).Send()

		return nil, err
	}

	return &utils.BatchwriteArgs{
		WriteRequests: writeRequests,
		TableName:     "media",
		MaxBatchSize:  25,
	}, nil
}
