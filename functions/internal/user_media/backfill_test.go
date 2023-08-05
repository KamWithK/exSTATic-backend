package user_media

import (
	"strings"
	"testing"

	"github.com/KamWithK/exSTATic-backend/internal/dynamo_wrapper"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

const EndpointURL = "http://localhost:4566/"

var sess *session.Session
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

func TestNull(t *testing.T) {
	result, err := PutBackfill(BackfillArgs{})

	assert.Nil(t, result, "No input => no writes")
	assert.Error(t, err)
}

func TestNoUsername(t *testing.T) {
	fake := faker.New()

	result, err := PutBackfill(BackfillArgs{
		Username:     "",
		MediaEntries: RandomMediaEntries(fake, "", 3),
	})

	assert.Nil(t, result, "Writes can't be performed when no username is entered")
	assert.Error(t, err)
}

func TestMultipleUsernames(t *testing.T) {
	fake := faker.New()
	user1 := fake.Person().Name()
	user2 := fake.Person().Name()

	invalidEntries, validEntries := 5, 3

	results, err := PutBackfill(BackfillArgs{
		Username:     user1,
		MediaEntries: append(RandomMediaEntries(fake, user2, invalidEntries), RandomMediaEntries(fake, user1, validEntries)...),
	})

	assert.Len(t, results.WriteRequests, validEntries, "Entries with different usernames aren't valid")
	assert.NoError(t, err)
}

func TestWriteMediaEntries(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	inputMediaEntries, producedMediaEntries := RandomMediaEntries(fake, user, 100), []UserMediaEntry{}

	results, err := PutBackfill(BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})

	assert.NoError(t, err)
	assert.NotNil(t, results)

	for _, writeRequest := range results.WriteRequests {
		intermediateItem := IntermediateEntryItem{}

		unmarshalErr := dynamodbattribute.UnmarshalMap(writeRequest.PutRequest.Item, &intermediateItem)
		assert.NoError(t, unmarshalErr)

		splitPK := strings.Split(intermediateItem.PK, "#")
		assert.Len(t, splitPK, 2, "PK should precisely be composed of the media type and username")
		intermediateItem.Key = UserMediaKey{
			Username:        splitPK[1],
			MediaType:       splitPK[0],
			MediaIdentifier: intermediateItem.SK,
		}

		producedMediaEntries = append(producedMediaEntries, intermediateItem.UserMediaEntry)
	}

	assert.Equal(t, inputMediaEntries, producedMediaEntries)
}

func TestWriteMediaStats(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	mediaEntries := RandomMediaEntries(fake, user, 100)
	inputMediaStats, producedMediaStats := []UserMediaStat{}, []UserMediaStat{}

	for _, mediaEntry := range mediaEntries {
		inputMediaStats = append(inputMediaStats, RandomMediaStats(fake, mediaEntry.Key, 30, 0.8)...)
	}

	results, err := PutBackfill(BackfillArgs{
		Username:   user,
		MediaStats: inputMediaStats,
	})

	assert.NoError(t, err)
	assert.NotNil(t, results)

	for _, writeRequest := range results.WriteRequests {
		intermediateItem := IntermediateStatItem{}

		unmarshalErr := dynamodbattribute.UnmarshalMap(writeRequest.PutRequest.Item, &intermediateItem)
		assert.NoError(t, unmarshalErr)

		splitPK := strings.Split(intermediateItem.PK, "#")
		splitSK := strings.Split(intermediateItem.SK, "#")
		assert.Len(t, splitPK, 2, "PK should precisely be composed of the media type and username")
		assert.Len(t, splitSK, 2, "SK should precisely be composed of the zero padded unix epoch date and media identifier")
		intermediateItem.Key = UserMediaKey{
			Username:        splitPK[1],
			MediaType:       splitPK[0],
			MediaIdentifier: splitSK[1],
		}

		producedMediaStats = append(producedMediaStats, intermediateItem.UserMediaStat)
	}

	assert.Equal(t, inputMediaStats, producedMediaStats)
}

func TestStorageRetrieval(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()
	numDays := 100

	inputMediaEntries := RandomMediaEntries(fake, user, numDays)
	batchwriterArgs, err := PutBackfill(BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, batchwriterArgs.WriteRequests)

	output := dynamo_wrapper.DistributedBatchWrites(dynamoSvc, batchwriterArgs)
	assert.Empty(t, output.WriteRequests)

	storedHistory, backfillErr := GetBackfill(dynamoSvc, UserMediaDateKey{
		Key: UserMediaKey{
			Username:  user,
			MediaType: "vn",
		},
		DateTime: 0,
	})
	assert.NoError(t, backfillErr)
	assert.NotEmpty(t, storedHistory.MediaEntries)

	newEntries := map[UserMediaKey]UserMediaEntry{}
	for _, newEntry := range storedHistory.MediaEntries {
		newEntries[newEntry.Key] = newEntry
	}

	for _, original := range inputMediaEntries {
		assert.Equal(t, original, newEntries[original.Key])
	}
}
