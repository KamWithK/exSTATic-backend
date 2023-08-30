package backfill

import (
	"strconv"
	"strings"
	"testing"

	"github.com/KamWithK/exSTATic-backend/internal/dynamo_wrapper"
	"github.com/KamWithK/exSTATic-backend/internal/user_media"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

type IntermediateEntryItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
	user_media.UserMediaEntry
}

type IntermediateStatItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
	user_media.UserMediaStat
}

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
		MediaEntries: user_media.RandomMediaEntries(fake, "", 3),
	})

	assert.Nil(t, result, "Writes can't be performed when no username is entered")
	assert.Error(t, err)
}

func TestMultipleUsernames(t *testing.T) {
	fake := faker.New()
	user1 := fake.Person().Name()
	user2 := fake.Person().Name()

	invalidEntries, validEntries := 5, 3

	userMediaEntries := user_media.RandomMediaEntries(fake, user1, validEntries)
	maps.Copy(userMediaEntries, user_media.RandomMediaEntries(fake, user2, invalidEntries))

	results, err := PutBackfill(BackfillArgs{
		Username:     user1,
		MediaEntries: userMediaEntries,
	})

	assert.Len(t, results.WriteRequests, validEntries, "Entries with different usernames aren't valid")
	assert.NoError(t, err)
}

func TestWriteMediaEntries(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	inputMediaEntries, producedMediaEntries := user_media.RandomMediaEntries(fake, user, 100), map[user_media.UserMediaKey]user_media.UserMediaEntry{}

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

		key := user_media.UserMediaKey{
			Username:        splitPK[1],
			MediaType:       splitPK[0],
			MediaIdentifier: intermediateItem.SK,
		}
		producedMediaEntries[key] = intermediateItem.UserMediaEntry
	}

	assert.Equal(t, inputMediaEntries, producedMediaEntries)
}

func TestWriteMediaStats(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	mediaEntries := user_media.RandomMediaEntries(fake, user, 100)
	inputMediaStats, producedMediaStats := map[user_media.UserMediaDateKey]user_media.UserMediaStat{}, map[user_media.UserMediaDateKey]user_media.UserMediaStat{}

	for key := range mediaEntries {
		maps.Copy(inputMediaStats, user_media.RandomMediaStats(fake, key, 30, 0.8))
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

		date, parseIntErr := strconv.ParseInt(splitSK[0], 10, 64)
		assert.NoError(t, parseIntErr)

		key := user_media.UserMediaDateKey{
			Key: user_media.UserMediaKey{
				Username:        splitPK[1],
				MediaType:       splitPK[0],
				MediaIdentifier: splitSK[1],
			},
			DateTime: date,
		}
		producedMediaStats[key] = intermediateItem.UserMediaStat
	}

	assert.Equal(t, inputMediaStats, producedMediaStats)
}

func TestStorageRetrieval(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()
	numDays := 100

	inputMediaEntries := user_media.RandomMediaEntries(fake, user, numDays)
	batchwriterArgs, err := PutBackfill(BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, batchwriterArgs.WriteRequests)

	output := dynamo_wrapper.DistributedBatchWrites(dynamoSvc, batchwriterArgs)
	assert.Empty(t, output.WriteRequests)

	storedHistory, backfillErr := GetBackfill(dynamoSvc, user_media.UserMediaDateKey{
		Key: user_media.UserMediaKey{
			Username:  user,
			MediaType: "vn",
		},
		DateTime: 0,
	})
	assert.NoError(t, backfillErr)
	assert.NotEmpty(t, storedHistory.MediaEntries)

	newEntries := map[user_media.UserMediaKey]user_media.UserMediaEntry{}
	for key, newEntry := range storedHistory.MediaEntries {
		newEntries[key] = newEntry
	}

	for key, original := range inputMediaEntries {
		assert.Equal(t, original, newEntries[key])
	}
}
