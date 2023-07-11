package backfill

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	dynamo_types "github.com/KamWithK/exSTATic-backend"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

const EndpointURL = "http://localhost:4566/"

var sess *session.Session
var svc *dynamodb.DynamoDB
var lambdaSvc *lambda.Lambda

var backfillLambda *string

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Endpoint:    aws.String(EndpointURL),
			Region:      aws.String(endpoints.UsEast1RegionID),
			Credentials: credentials.NewStaticCredentials("foo", "var", ""),
		},
	}))
	// dynamoSvc = dynamodb.New(sess)
	lambdaSvc = lambda.New(sess)

	lambdas, err := lambdaSvc.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		return
	}

	for _, function := range lambdas.Functions {
		if strings.Contains(*function.FunctionName, "backfillPostFunction") {
			backfillLambda = function.FunctionName
			break
		}
	}
}

func RandomVNKey(fake faker.Faker, user string) dynamo_types.UserMediaKey {
	return dynamo_types.UserMediaKey{
		Username:        user,
		MediaType:       "vn",
		MediaIdentifier: fake.Directory().Directory(2),
	}
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

// Call function without data
func TestNull(t *testing.T) {
	input := &lambda.InvokeInput{
		FunctionName: backfillLambda,
		Payload:      []byte{},
	}

	result, err := lambdaSvc.Invoke(input)
	assert.NoError(t, err)
	assert.Equal(t, int64(200), *result.StatusCode)
}

func TestRandomBackfill(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	var mediaEntries []dynamo_types.UserMediaEntry
	var mediaStats []dynamo_types.UserMediaStat

	for i := 0; i < 5; i++ {
		key := RandomVNKey(fake, user)

		mediaEntries = append(mediaEntries, dynamo_types.UserMediaEntry{
			Key:         key,
			DisplayName: "",
			Series:      "",
			LastUpdate:  0,
		})
		mediaStats = append(mediaStats, RandomMediaStats(fake, key, 30, 0.8)...)
	}
}
