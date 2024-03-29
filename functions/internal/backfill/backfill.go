package backfill

import (
	"errors"

	"github.com/KamWithK/exSTATic-backend/internal/dynamo_wrapper"
	"github.com/KamWithK/exSTATic-backend/internal/user_media"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog/log"
)

type BackfillArgs struct {
	Username     string                                                   `json:"username"`
	MediaEntries map[user_media.UserMediaKey]user_media.UserMediaEntry    `json:"media_entries"`
	MediaStats   map[user_media.UserMediaDateKey]user_media.UserMediaStat `json:"media_stats"`
}

func GetBackfill(svc *dynamodb.DynamoDB, userMediaDateKey user_media.UserMediaDateKey) (*BackfillArgs, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("media"),
		KeyConditionExpression: aws.String("pk = :pk AND last_update >= :lastUpdate"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(user_media.UserMediaPK(userMediaDateKey.Key)),
			},
			":lastUpdate": {
				N: aws.String(user_media.ZeroPadInt64(userMediaDateKey.DateTime)),
			},
		},
		IndexName: aws.String("lastUpdatedIndex"),
	}

	result, queryErr := svc.Query(queryInput)
	if queryErr != nil {
		log.Info().Err(queryErr).Send()
		return nil, queryErr
	}

	mediaEntries := map[user_media.UserMediaKey]user_media.UserMediaEntry{}
	mediaStats := map[user_media.UserMediaDateKey]user_media.UserMediaStat{}

	for _, item := range result.Items {
		pk, sk := *item["pk"].S, *item["sk"].S
		key, date, splitErr := user_media.SplitUserMediaCompositeKey(pk, sk)

		if splitErr != nil {
			log.Error().Err(splitErr).Str("pk", pk).Str("sk", sk).Interface("item", item).Msg("Could not split keys")
		} else if key != nil && date == nil {
			mediaEntry := user_media.UserMediaEntry{}
			unmarshalErr := dynamodbattribute.UnmarshalMap(item, &mediaEntry)

			if unmarshalErr != nil {
				log.Error().Interface("key", key).Interface("item", item).Err(unmarshalErr).Msg("Could not unmarshal item into UserMediaEntry")
			} else {
				mediaEntries[*key] = mediaEntry
			}
		} else if key != nil {
			mediaStat := user_media.UserMediaStat{}
			unmarshalErr := dynamodbattribute.UnmarshalMap(item, &mediaStat)

			if unmarshalErr != nil {
				log.Error().Interface("key", key).Interface("item", item).Err(unmarshalErr).Msg("Could not unmarshal item into UserMediaStat")
			} else {
				dateKey := user_media.UserMediaDateKey{
					Key:      *key,
					DateTime: *date,
				}
				mediaStats[dateKey] = mediaStat
			}
		} else {
			log.Error().Str("pk", pk).Str("sk", sk).Interface("item", item).Msg("Item neither entry nor stat")
		}
	}

	return &BackfillArgs{
		Username:     userMediaDateKey.Key.Username,
		MediaEntries: mediaEntries,
		MediaStats:   mediaStats,
	}, nil
}

func PutBackfill(history BackfillArgs) (*dynamo_wrapper.BatchwriteArgs, error) {
	username := history.Username

	if len(username) == 0 {
		err := errors.New("invalid username")
		log.Info().Err(err).Send()

		return nil, err
	}

	writeRequests := []*dynamodb.WriteRequest{}

	for key, userMedia := range history.MediaEntries {
		writeRequest := dynamo_wrapper.PutRawRequest(user_media.UserMediaPK(key), user_media.MediaInfoSK(key), &userMedia)
		if key.Username != username {
			err := errors.New("username mismatch")
			log.Info().Err(err).Send()
		} else if writeRequest != nil {
			writeRequests = append(writeRequests, writeRequest)
		}
	}

	for key, userMedia := range history.MediaStats {
		writeRequest := dynamo_wrapper.PutRawRequest(user_media.UserMediaPK(key.Key), user_media.StatusUpdateSK(key), &userMedia)
		if key.Key.Username != username {
			err := errors.New("username mismatch")
			log.Info().Err(err).Send()
		} else if writeRequest != nil {
			writeRequests = append(writeRequests, writeRequest)
		}
	}

	if len(writeRequests) == 0 {
		err := errors.New("error no valid data")
		log.Info().Err(err).Send()

		return nil, err
	}

	return &dynamo_wrapper.BatchwriteArgs{
		WriteRequests: writeRequests,
		TableName:     "media",
		MaxBatchSize:  25,
	}, nil
}
