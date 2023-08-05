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
	"github.com/jaswdr/faker"

	"github.com/aws/aws-sdk-go/service/dynamodb"
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

func TestWriteMediaEntries(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	inputMediaEntries := random_data.RandomMediaEntries(fake, user, 100)
	batchwriterArgs, err := user_media.PutBackfill(user_media.BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, batchwriterArgs.WriteRequests)

	output := utils.DistributedBatchWrites(dynamoSvc, batchwriterArgs)
	assert.Empty(t, output.WriteRequests)

	for _, original := range inputMediaEntries {
		result, err := user_media.GetMediaInfo(dynamoSvc, original.Key)
		assert.NoError(t, err)
		assert.Equal(t, original, *result)
	}
}
