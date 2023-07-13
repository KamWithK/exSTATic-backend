package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/KamWithK/exSTATic-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jaswdr/faker"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/stretchr/testify/assert"
)

const EndpointURL = "http://localhost:4566/"

var lambdaSvc *lambda.Lambda
var dynamoSvc *dynamodb.DynamoDB

var backfillLambda *string
var mediaInfoGetLambda *string

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
	lambdaSvc = lambda.New(sess)

	lambdas, err := lambdaSvc.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		return
	}

	for _, function := range lambdas.Functions {
		if strings.Contains(*function.FunctionName, "backfillPostFunction") {
			backfillLambda = function.FunctionName
			if mediaInfoGetLambda != nil {
				break
			}
		} else if strings.Contains(*function.FunctionName, "mediaInfoGetFunction") {
			mediaInfoGetLambda = function.FunctionName
			if backfillLambda != nil {
				break
			}
		}
	}
}

func DynamoEquality(t *testing.T, original utils.UserMediaEntry) {
	payload, marshalErr := json.Marshal(original.Key)
	assert.NoError(t, marshalErr)

	input := &lambda.InvokeInput{
		FunctionName: mediaInfoGetLambda,
		Payload:      payload,
	}

	result, err := lambdaSvc.Invoke(input)
	assert.NoError(t, err)
	assert.Equal(t, int64(200), *result.StatusCode)

	log.Error().Interface("og key", original.Key).Str("error", *result.FunctionError).Msg("Idk")
	assert.Nil(t, result.FunctionError)

	var output utils.UserMediaEntry
	unmarshalErr := json.Unmarshal(result.Payload, &output)
	assert.NoError(t, unmarshalErr)
	assert.Equal(t, original, output)
}

func RandomMediaEntryWrites(t *testing.T, args utils.BackfillArgs) *utils.BatchwriteArgs {
	payload, marshalErr := json.Marshal(args)
	assert.NoError(t, marshalErr)

	input := &lambda.InvokeInput{
		FunctionName: backfillLambda,
		Payload:      payload,
	}

	result, err := lambdaSvc.Invoke(input)
	assert.NoError(t, err)
	assert.Equal(t, int64(200), *result.StatusCode)

	assert.Nil(t, result.FunctionError)

	var output *utils.BatchwriteArgs
	unmarshalErr := json.Unmarshal(result.Payload, &output)
	assert.NoError(t, unmarshalErr)
	assert.NotEmpty(t, output.WriteRequests)

	return output
}

func TestWriteMediaEntries(t *testing.T) {
	fake := faker.New()
	user := fake.Person().Name()

	inputMediaEntries := utils.RandomMediaEntries(fake, user, 1)
	batchwriterArgs := RandomMediaEntryWrites(t, utils.BackfillArgs{
		Username:     user,
		MediaEntries: inputMediaEntries,
	})
	_ = batchwriterArgs

	output := utils.DistributedBatchWrites(dynamoSvc, batchwriterArgs)
	assert.Empty(t, output.WriteRequests)

	for _, original := range inputMediaEntries {
		DynamoEquality(t, original)
	}
}
