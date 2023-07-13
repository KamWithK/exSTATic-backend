package backfill

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
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
