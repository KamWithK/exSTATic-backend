package models

import (
	"errors"

	"github.com/KamWithK/exSTATic-backend/utils"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/rs/zerolog/log"
)

type BackfillArgs struct {
	Username     string           `json:"username"`
	MediaEntries []UserMediaEntry `json:"media_entries"`
	MediaStats   []UserMediaStat  `json:"media_stats"`
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
