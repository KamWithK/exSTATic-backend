package main

import (
	"strings"
	"testing"

	"github.com/KamWithK/exSTATic-backend/utils"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

func TestNull(t *testing.T) {
	result, err := HandleRequest(nil, utils.BackfillArgs{})

	assert.Nil(t, result, "No input => no writes")
	assert.Error(t, err)
}

func TestNoUsername(t *testing.T) {
	fake := faker.New()

	result, err := HandleRequest(nil, utils.BackfillArgs{
		Username:     "",
		MediaEntries: utils.RandomMediaEntries(fake, "", 3),
	})

	assert.Nil(t, result, "Writes can't be performed when no username is entered")
	assert.Error(t, err)
}

func TestMultipleUsernames(t *testing.T) {
	fake := faker.New()
	user1 := fake.Person().Name()
	user2 := fake.Person().Name()

	invalidEntries, validEntries := 5, 3

	results, err := HandleRequest(nil, utils.BackfillArgs{
		Username:     user1,
		MediaEntries: append(utils.RandomMediaEntries(fake, user2, invalidEntries), utils.RandomMediaEntries(fake, user1, validEntries)...),
	})

	assert.Len(t, results.WriteRequests, validEntries, "Entries with different usernames aren't valid")
	assert.NoError(t, err)
}

func TestWriteMediaEntries(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	inputMediaEntries, producedMediaEntries := utils.RandomMediaEntries(fake, user, 100), []utils.UserMediaEntry{}

	results, err := HandleRequest(nil, utils.BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})

	assert.NoError(t, err)
	assert.NotNil(t, results)

	for _, writeRequest := range results.WriteRequests {
		intermediateItem := utils.IntermediateEntryItem{}

		unmarshalErr := dynamodbattribute.UnmarshalMap(writeRequest.PutRequest.Item, &intermediateItem)
		assert.NoError(t, unmarshalErr)

		splitPK := strings.Split(intermediateItem.PK, "#")
		assert.Len(t, splitPK, 2, "PK should precisely be composed of the media type and username")
		intermediateItem.Key = utils.UserMediaKey{
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

	mediaEntries := utils.RandomMediaEntries(fake, user, 100)
	inputMediaStats, producedMediaStats := []utils.UserMediaStat{}, []utils.UserMediaStat{}

	for _, mediaEntry := range mediaEntries {
		inputMediaStats = append(inputMediaStats, utils.RandomMediaStats(fake, mediaEntry.Key, 30, 0.8)...)
	}

	results, err := HandleRequest(nil, utils.BackfillArgs{
		Username:   user,
		MediaStats: inputMediaStats,
	})

	assert.NoError(t, err)
	assert.NotNil(t, results)

	for _, writeRequest := range results.WriteRequests {
		intermediateItem := utils.IntermediateStatItem{}

		unmarshalErr := dynamodbattribute.UnmarshalMap(writeRequest.PutRequest.Item, &intermediateItem)
		assert.NoError(t, unmarshalErr)

		splitPK := strings.Split(intermediateItem.PK, "#")
		splitSK := strings.Split(intermediateItem.SK, "#")
		assert.Len(t, splitPK, 2, "PK should precisely be composed of the media type and username")
		assert.Len(t, splitSK, 2, "SK should precisely be composed of the zero padded unix epoch date and media identifier")
		intermediateItem.Key = utils.UserMediaKey{
			Username:        splitPK[1],
			MediaType:       splitPK[0],
			MediaIdentifier: splitSK[1],
		}

		producedMediaStats = append(producedMediaStats, intermediateItem.UserMediaStat)
	}

	assert.Equal(t, inputMediaStats, producedMediaStats)
}
