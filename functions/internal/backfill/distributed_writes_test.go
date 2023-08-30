package backfill

import (
	"testing"

	"github.com/KamWithK/exSTATic-backend/internal/dynamo_wrapper"
	"github.com/KamWithK/exSTATic-backend/internal/user_media"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jaswdr/faker"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
)

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

func TestDistributedBatchWrites(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	inputMediaEntries := user_media.RandomMediaEntries(fake, user, 100)
	batchwriterArgs, err := PutBackfill(BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, batchwriterArgs.WriteRequests)

	output := dynamo_wrapper.DistributedBatchWrites(dynamoSvc, batchwriterArgs)
	assert.Empty(t, output.WriteRequests)

	for key, original := range inputMediaEntries {
		result, err := user_media.GetMediaInfo(dynamoSvc, key)
		assert.NoError(t, err)
		assert.Equal(t, original, *result)
	}
}
