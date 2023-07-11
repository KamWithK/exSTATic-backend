package backfill

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/stretchr/testify/assert"
)

const EndpointURL = "http://localhost:4566/"

var sess *session.Session
var svc *dynamodb.DynamoDB
var lambdaSvc *lambda.Lambda

var backfillLambda *string

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		// SharedConfigState: session.SharedConfigEnable,
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
		fmt.Println("~~~~~~~~~~~~~~~~~Failed to list functions:", err)
		return
	}

	fmt.Println("Found", len(lambdas.Functions), "functions")

	for _, function := range lambdas.Functions {
		fmt.Println("Found function: ", *function.FunctionName)

		if strings.Contains(*function.FunctionName, "backfillPostFunction") {
			backfillLambda = function.FunctionName
			break
		}
	}
}

func TestBackfill(t *testing.T) {
	input := &lambda.InvokeInput{
		FunctionName: backfillLambda,
		Payload:      []byte{},
	}

	result, err := lambdaSvc.Invoke(input)
	assert.NoError(t, err)
	assert.Equal(t, int64(200), *result.StatusCode)
}
