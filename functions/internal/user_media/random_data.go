package user_media

import (
	"math/rand"
	"time"

	"github.com/jaswdr/faker"
)

func RandomVNKey(fake faker.Faker, user string) UserMediaKey {
	return UserMediaKey{
		Username:        user,
		MediaType:       "vn",
		MediaIdentifier: fake.Directory().Directory(2),
	}
}

func RandomMediaEntries(fake faker.Faker, user string, numEntries int) map[UserMediaKey]UserMediaEntry {
	mediaEntries := map[UserMediaKey]UserMediaEntry{}

	for i := 0; i < numEntries; i++ {
		key := RandomVNKey(fake, user)

		mediaEntries[key] = UserMediaEntry{
			DisplayName: fake.RandomLetter(),
			Series:      fake.RandomLetter(),
			LastUpdate:  0,
		}
	}

	return mediaEntries
}

// Create a random stats entry for some number of days in the past
func RandomMediaStats(fake faker.Faker, key UserMediaKey, daysAgo int, probability float32) map[UserMediaDateKey]UserMediaStat {
	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -1*daysAgo)

	stats := map[UserMediaDateKey]UserMediaStat{}

	for day := startDate; day.Before(now) || day.Equal(now); day = day.AddDate(0, 0, 1) {
		if rand.Float32() < probability {
			dateKey := UserMediaDateKey{
				Key:      key,
				DateTime: day.Unix(),
			}
			stats[dateKey] = UserMediaStat{
				Stats: MediaStat{
					TimeRead:  fake.Int64Between(1000, 5000),
					CharsRead: fake.Int64Between(100, 5000),
					LinesRead: fake.Int64Between(0, 500),
				},
				Pause: false,
			}
		}
	}

	return stats
}
