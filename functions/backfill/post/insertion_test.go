package main

import (
	"math/rand"
	"testing"
	"time"

	dynamo_types "github.com/KamWithK/exSTATic-backend"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

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
	user := fake.Person().Name()

	result, err := HandleRequest(nil, BackfillArgs{
		Username:     "",
		MediaEntries: RandomMediaEntries(fake, user, 3),
	})

	assert.Nil(t, result, "Writes can't be performed when no username is entered")
	assert.Error(t, err)
}

func TestWriteMediaEntries(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	results, err := HandleRequest(nil, BackfillArgs{
		Username:     user,
		MediaEntries: RandomMediaEntries(fake, user, 100),
	})

	assert.NoError(t, err)
	assert.NotNil(t, results)
}

func TestWriteMediaStats(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	mediaEntries := RandomMediaEntries(fake, user, 100)
	var mediaStats []dynamo_types.UserMediaStat

	for _, mediaEntry := range mediaEntries {
		mediaStats = append(mediaStats, RandomMediaStats(fake, mediaEntry.Key, 30, 0.8)...)
	}

	results, err := HandleRequest(nil, BackfillArgs{
		Username:   user,
		MediaStats: mediaStats,
	})

	assert.NoError(t, err)
	assert.NotNil(t, results)
}
