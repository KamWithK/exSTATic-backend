package main

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	dynamo_types "github.com/KamWithK/exSTATic-backend"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

type IntermediateEntryItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
	dynamo_types.UserMediaEntry
}

type IntermediateStatItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
	dynamo_types.UserMediaStat
}

func RandomVNKey(fake faker.Faker, user string) dynamo_types.UserMediaKey {
	return dynamo_types.UserMediaKey{
		Username:        user,
		MediaType:       "vn",
		MediaIdentifier: fake.Directory().Directory(2),
	}
}

func RandomMediaEntries(fake faker.Faker, user string, numEntries int) []dynamo_types.UserMediaEntry {
	var mediaEntries []dynamo_types.UserMediaEntry

	for i := 0; i < numEntries; i++ {
		key := RandomVNKey(fake, user)

		mediaEntries = append(mediaEntries, dynamo_types.UserMediaEntry{
			Key:         key,
			DisplayName: fake.RandomLetter(),
			Series:      fake.RandomLetter(),
			LastUpdate:  0,
		})
	}

	return mediaEntries
}

// Create a random stats entry for some number of days in the past
func RandomMediaStats(fake faker.Faker, key dynamo_types.UserMediaKey, daysAgo int, probability float32) []dynamo_types.UserMediaStat {
	now := time.Now()
	startDate := now.AddDate(0, 0, -1*daysAgo)

	var stats []dynamo_types.UserMediaStat

	for day := startDate; day.Before(now) || day.Equal(now); day = day.AddDate(0, 0, 1) {
		if rand.Float32() < probability {
			stats = append(stats, dynamo_types.UserMediaStat{
				Key:  key,
				Date: aws.Int64(day.Unix()),
				Stats: dynamo_types.MediaStat{
					TimeRead:  fake.Int64Between(1000, 5000),
					CharsRead: fake.Int64Between(100, 5000),
					LinesRead: fake.Int64Between(0, 500),
				},
				Pause: false,
			})
		}
	}

	return stats
}

func TestNull(t *testing.T) {
	result, err := HandleRequest(nil, BackfillArgs{})

	assert.Nil(t, result, "No input => no writes")
	assert.Error(t, err)
}

func TestNoUsername(t *testing.T) {
	fake := faker.New()

	result, err := HandleRequest(nil, BackfillArgs{
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

	results, err := HandleRequest(nil, BackfillArgs{
		Username:     user1,
		MediaEntries: append(RandomMediaEntries(fake, user2, invalidEntries), RandomMediaEntries(fake, user1, validEntries)...),
	})

	assert.Len(t, results.WriteRequests, validEntries, "Entries with different usernames aren't valid")
	assert.NoError(t, err)
}

func TestWriteMediaEntries(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	inputMediaEntries, producedMediaEntries := RandomMediaEntries(fake, user, 100), []dynamo_types.UserMediaEntry{}

	results, err := HandleRequest(nil, BackfillArgs{
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
		intermediateItem.Key = dynamo_types.UserMediaKey{
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
	inputMediaStats, producedMediaStats := []dynamo_types.UserMediaStat{}, []dynamo_types.UserMediaStat{}

	for _, mediaEntry := range mediaEntries {
		inputMediaStats = append(inputMediaStats, RandomMediaStats(fake, mediaEntry.Key, 30, 0.8)...)
	}

	results, err := HandleRequest(nil, BackfillArgs{
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
		intermediateItem.Key = dynamo_types.UserMediaKey{
			Username:        splitPK[1],
			MediaType:       splitPK[0],
			MediaIdentifier: splitSK[1],
		}

		producedMediaStats = append(producedMediaStats, intermediateItem.UserMediaStat)
	}

	assert.Equal(t, inputMediaStats, producedMediaStats)
}
