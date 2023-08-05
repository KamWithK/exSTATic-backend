package main

import (
	"strings"
	"testing"

	"github.com/KamWithK/exSTATic-backend/internal/random_data"
	"github.com/KamWithK/exSTATic-backend/internal/user_media"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

func TestNull(t *testing.T) {
	result, err := user_media.PutBackfill(user_media.BackfillArgs{})

	assert.Nil(t, result, "No input => no writes")
	assert.Error(t, err)
}

func TestNoUsername(t *testing.T) {
	fake := faker.New()

	result, err := user_media.PutBackfill(user_media.BackfillArgs{
		Username:     "",
		MediaEntries: random_data.RandomMediaEntries(fake, "", 3),
	})

	assert.Nil(t, result, "Writes can't be performed when no username is entered")
	assert.Error(t, err)
}

func TestMultipleUsernames(t *testing.T) {
	fake := faker.New()
	user1 := fake.Person().Name()
	user2 := fake.Person().Name()

	invalidEntries, validEntries := 5, 3

	results, err := user_media.PutBackfill(user_media.BackfillArgs{
		Username:     user1,
		MediaEntries: append(random_data.RandomMediaEntries(fake, user2, invalidEntries), random_data.RandomMediaEntries(fake, user1, validEntries)...),
	})

	assert.Len(t, results.WriteRequests, validEntries, "Entries with different usernames aren't valid")
	assert.NoError(t, err)
}

func TestWriteMediaEntries(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	inputMediaEntries, producedMediaEntries := random_data.RandomMediaEntries(fake, user, 100), []user_media.UserMediaEntry{}

	results, err := user_media.PutBackfill(user_media.BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})

	assert.NoError(t, err)
	assert.NotNil(t, results)

	for _, writeRequest := range results.WriteRequests {
		intermediateItem := user_media.IntermediateEntryItem{}

		unmarshalErr := dynamodbattribute.UnmarshalMap(writeRequest.PutRequest.Item, &intermediateItem)
		assert.NoError(t, unmarshalErr)

		splitPK := strings.Split(intermediateItem.PK, "#")
		assert.Len(t, splitPK, 2, "PK should precisely be composed of the media type and username")
		intermediateItem.Key = user_media.UserMediaKey{
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

	mediaEntries := random_data.RandomMediaEntries(fake, user, 100)
	inputMediaStats, producedMediaStats := []user_media.UserMediaStat{}, []user_media.UserMediaStat{}

	for _, mediaEntry := range mediaEntries {
		inputMediaStats = append(inputMediaStats, random_data.RandomMediaStats(fake, mediaEntry.Key, 30, 0.8)...)
	}

	results, err := user_media.PutBackfill(user_media.BackfillArgs{
		Username:   user,
		MediaStats: inputMediaStats,
	})

	assert.NoError(t, err)
	assert.NotNil(t, results)

	for _, writeRequest := range results.WriteRequests {
		intermediateItem := user_media.IntermediateStatItem{}

		unmarshalErr := dynamodbattribute.UnmarshalMap(writeRequest.PutRequest.Item, &intermediateItem)
		assert.NoError(t, unmarshalErr)

		splitPK := strings.Split(intermediateItem.PK, "#")
		splitSK := strings.Split(intermediateItem.SK, "#")
		assert.Len(t, splitPK, 2, "PK should precisely be composed of the media type and username")
		assert.Len(t, splitSK, 2, "SK should precisely be composed of the zero padded unix epoch date and media identifier")
		intermediateItem.Key = user_media.UserMediaKey{
			Username:        splitPK[1],
			MediaType:       splitPK[0],
			MediaIdentifier: splitSK[1],
		}

		producedMediaStats = append(producedMediaStats, intermediateItem.UserMediaStat)
	}

	assert.Equal(t, inputMediaStats, producedMediaStats)
}
