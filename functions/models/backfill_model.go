package models

import (
	"errors"
	"strconv"
	"strings"

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

func SplitCompositeKey(pk string, sk string) (*UserMediaKey, *int64, error) {
	if pk == "" {
		return nil, nil, errors.New("Empty partition key (pk)")
	}
	if sk == "" {
		return nil, nil, errors.New("Empty secondary key (sk)")
	}

	pkSplit := strings.Split(pk, "#")
	skSplit := strings.Split(sk, "#")

	key := UserMediaKey{}
	var recordDate *int64

	if len(pkSplit) != 2 || pkSplit[0] == "" || pkSplit[1] == "" {
		return nil, nil, errors.New("Invalid partition key (pk) split")
	}

	key.Username = pkSplit[1]
	key.MediaType = pkSplit[0]

	if len(skSplit) == 1 {
		key.MediaIdentifier = skSplit[0]
	} else if len(skSplit) == 2 {
		key.MediaIdentifier = skSplit[1]
		date, parseIntErr := strconv.ParseInt(skSplit[0], 10, 64)

		if parseIntErr != nil {
			return nil, nil, errors.New("Could not parse Unix epoch")
		}

		recordDate = &date
	} else {
		return nil, nil, errors.New("Invalid secondary key (sk) split")
	}

	return &key, recordDate, nil
}

func GetBackfill(svc *dynamodb.DynamoDB, UserMediaDateKey UserMediaDateKey) (*BackfillArgs, error) {
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

	mediaEntries := []UserMediaEntry{}
	mediaStats := []UserMediaStat{}

	for _, item := range result.Items {
		pk, sk := item["pk"].String(), item["sk"].String()
		key, date, splitErr := SplitCompositeKey(pk, sk)

		if splitErr != nil {
			log.Error().Err(splitErr).Str("pk", pk).Str("sk", sk).Interface("item", item).Msg("Could not split keys")
		} else if date == nil {
			mediaEntry := UserMediaEntry{}
			unmarshalErr := dynamodbattribute.UnmarshalMap(item, &mediaEntry)
			if unmarshalErr != nil {
				log.Error().Interface("key", key).Interface("item", item).Err(unmarshalErr).Msg("Could not unmarshal item into UserMediaEntry")
			}
		} else if key != nil {
			mediaStat := UserMediaStat{}
			unmarshalErr := dynamodbattribute.UnmarshalMap(item, &mediaStat)
			if unmarshalErr != nil {
				log.Error().Interface("key", key).Interface("item", item).Err(unmarshalErr).Msg("Could not unmarshal item into UserMediaStat")
			}
		} else {
			log.Error().Str("pk", pk).Str("sk", sk).Interface("item", item).Msg("Item neither entry nor stat")
		}
	}

	return &BackfillArgs{
		Username:     UserMediaDateKey.Key.Username,
		MediaEntries: mediaEntries,
		MediaStats:   mediaStats,
	}, nil
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
