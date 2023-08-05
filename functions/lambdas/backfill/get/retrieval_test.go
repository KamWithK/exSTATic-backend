package main

import (
	"testing"

	"github.com/KamWithK/exSTATic-backend/internal/models"
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

	inputMediaEntries := models.RandomMediaEntries(fake, user, numDays)
	batchwriterArgs, err := models.PutBackfill(models.BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, batchwriterArgs.WriteRequests)

	output := utils.DistributedBatchWrites(dynamoSvc, batchwriterArgs)
	assert.Empty(t, output.WriteRequests)

	storedHistory, backfillErr := models.GetBackfill(dynamoSvc, models.UserMediaDateKey{
		Key: models.UserMediaKey{
			Username:  user,
			MediaType: "vn",
		},
		DateTime: 0,
	})
	assert.NoError(t, backfillErr)
	assert.NotEmpty(t, storedHistory.MediaEntries)

	newEntries := map[models.UserMediaKey]models.UserMediaEntry{}
	for _, newEntry := range storedHistory.MediaEntries {
		newEntries[newEntry.Key] = newEntry
	}

	for _, original := range inputMediaEntries {
		assert.Equal(t, original, newEntries[original.Key])
	}
}
