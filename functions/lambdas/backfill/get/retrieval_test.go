package main

import (
	"testing"

	"github.com/KamWithK/exSTATic-backend/internal/random_data"
	"github.com/KamWithK/exSTATic-backend/internal/user_media"
	"github.com/KamWithK/exSTATic-backend/internal/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

const EndpointURL = "http://localhost:4566/"

var dynamoSvc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Endpoint:    aws.String(EndpointURL),
			Region:      aws.String(endpoints.UsEast1RegionID),
			Credentials: credentials.NewStaticCredentials("foo", "var", ""),
		},
		SharedConfigState: session.SharedConfigEnable,
	}))
	dynamoSvc = dynamodb.New(sess)
}

func TestBackfillEntries(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()
	numDays := 100

	inputMediaEntries := random_data.RandomMediaEntries(fake, user, numDays)
	batchwriterArgs, err := user_media.PutBackfill(user_media.BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, batchwriterArgs.WriteRequests)

	output := utils.DistributedBatchWrites(dynamoSvc, batchwriterArgs)
	assert.Empty(t, output.WriteRequests)

	storedHistory, backfillErr := user_media.GetBackfill(dynamoSvc, user_media.UserMediaDateKey{
		Key: user_media.UserMediaKey{
			Username:  user,
			MediaType: "vn",
		},
		DateTime: 0,
	})
	assert.NoError(t, backfillErr)
	assert.NotEmpty(t, storedHistory.MediaEntries)

	newEntries := map[user_media.UserMediaKey]user_media.UserMediaEntry{}
	for _, newEntry := range storedHistory.MediaEntries {
		newEntries[newEntry.Key] = newEntry
	}

	for _, original := range inputMediaEntries {
		assert.Equal(t, original, newEntries[original.Key])
	}
}
