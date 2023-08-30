package user_media

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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

func TestGetNoDay(t *testing.T) {
	key := UserMediaDateKey{
		Key: UserMediaKey{
			Username:        "username",
			MediaType:       "vn",
			MediaIdentifier: "identifier",
		},
		DateTime: 0,
	}

	deleteErr := DeleteStatusUpdate(dynamoSvc, key)
	assert.NoError(t, deleteErr)

	_, userMediaStats, findDayErr := GetStatusUpdate(dynamoSvc, key)

	assert.Error(t, findDayErr, ErrEmptyItems)
	assert.Empty(t, userMediaStats.Stats)
}

func TestPutMediaInfo(t *testing.T) {
	key := UserMediaKey{
		Username:        "username",
		MediaType:       "vn",
		MediaIdentifier: "identifier",
	}

	error := PutMediaInfo(dynamoSvc, key, UserMediaEntry{
		DisplayName: "name",
	}, 0)
	assert.NoError(t, error)
}

func TestBasicInsert(t *testing.T) {
	fake := faker.New()
	key := UserMediaDateKey{
		Key: UserMediaKey{
			Username:        "username",
			MediaType:       "vn",
			MediaIdentifier: "identifier",
		},
		DateTime: 0,
	}
	additiveStat := MediaStat{
		CharsRead: 1000,
		LinesRead: 1000,
	}
	var maxAFKTime int16 = 60

	_, oldUserMediaStats, _ := GetStatusUpdate(dynamoSvc, key)

	deleteErr := DeleteStatusUpdate(dynamoSvc, key)
	assert.NoError(t, deleteErr)

	putErr := PutStatusUpdate(dynamoSvc, StatusArgs{
		Key:      key.Key,
		Stats:    additiveStat,
		Progress: make(ProgressPoints, 1),
		Timezone: fake.Time().Timezone(),
	}, maxAFKTime)
	assert.NoError(t, putErr)

	_, userMediaStats, findDayErr := GetStatusUpdate(dynamoSvc, key)

	assert.NoError(t, findDayErr)
	assert.Equal(t, userMediaStats.Stats.CharsRead, oldUserMediaStats.Stats.CharsRead+additiveStat.CharsRead)
	assert.Equal(t, userMediaStats.Stats.LinesRead, oldUserMediaStats.Stats.LinesRead+additiveStat.LinesRead)
}
