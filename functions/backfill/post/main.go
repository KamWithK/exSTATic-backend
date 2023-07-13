package main

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/KamWithK/exSTATic-backend/utils"
)

func putRequest(pk string, sk string, userMedia interface{}) *dynamodb.WriteRequest {
	tableKey, keyErr := utils.GetCompositeKey(pk, sk)
	if keyErr != nil {
		return nil
	}

	writeRequest, writeErr := utils.PutItemRequest(tableKey, userMedia)

	if writeErr != nil {
		return nil
	}

	return writeRequest
}

func HandleRequest(ctx context.Context, history utils.BackfillArgs) (*utils.BatchwriteArgs, error) {
	username := history.Username

	if len(username) == 0 {
		err := errors.New("Invalid username")
		log.Info().Err(err).Send()

		return nil, err
	}

	writeRequests := []*dynamodb.WriteRequest{}

	for _, userMedia := range history.MediaEntries {
		writeRequest := putRequest(userMedia.Key.MediaType+"#"+username, userMedia.Key.MediaIdentifier, &userMedia)
		if userMedia.Key.Username != username {
			err := errors.New("Username mismatch")
			log.Info().Err(err).Send()
		} else if writeRequest != nil {
			writeRequests = append(writeRequests, writeRequest)
		}
	}

	for _, userMedia := range history.MediaStats {
		writeRequest := putRequest(userMedia.Key.MediaType+"#"+username, utils.ZeroPadInt64(*userMedia.Date)+"#"+userMedia.Key.MediaIdentifier, &userMedia)
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

func main() {
	lambda.Start(HandleRequest)
}
