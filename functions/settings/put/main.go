package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	dynamo_types "github.com/KamWithK/exSTATic-backend"
)

var sess *session.Session
var svc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, options dynamo_types.UserSettings) error {
	key := dynamo_types.UserSettingsKey{
		Username:  options.Username,
		MediaType: options.MediaType,
	}
	tableKey, keyErr := dynamodbattribute.MarshalMap(key)

	if keyErr != nil {
		return fmt.Errorf("Error marshalling key: %s", keyErr.Error())
	}

	updateExpression, expressionAttributeNames, expressionAttributeValues := createUpdateExpressionAttributes(options)

	_, updateErr := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String("settings"),
		Key:                       tableKey,
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	})

	if updateErr != nil {
		return fmt.Errorf("Error updating DynamoDB item: %s", updateErr.Error())
	}

	return nil
}

func addAttributeIfNotNull(updateExpression string, expressionAttributeNames map[string]*string, expressionAttributeValues map[string]*dynamodb.AttributeValue, attributeName, jsonAttributeName string, value interface{}) (string, map[string]*string, map[string]*dynamodb.AttributeValue) {
	if value != nil {
		if len(expressionAttributeNames) > 0 {
			updateExpression += ","
		}
		updateExpression += " #" + attributeName + " = :" + attributeName
		expressionAttributeNames["#"+attributeName] = aws.String(jsonAttributeName)
		value, _ := dynamodbattribute.Marshal(value)
		expressionAttributeValues[":"+attributeName] = value
	}
	return updateExpression, expressionAttributeNames, expressionAttributeValues
}

func createUpdateExpressionAttributes(optionArgs dynamo_types.UserSettings) (string, map[string]*string, map[string]*dynamodb.AttributeValue) {
	updateExpression := "SET"
	expressionAttributeNames := map[string]*string{}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{}

	valueOfOptionArgs := reflect.ValueOf(optionArgs)
	typeOfOptionArgs := valueOfOptionArgs.Type()

	for i := 0; i < valueOfOptionArgs.NumField(); i++ {
		field := valueOfOptionArgs.Field(i)
		fieldType := typeOfOptionArgs.Field(i)

		if field.Kind() != reflect.Invalid && !field.IsZero() {
			jsonTag := strings.Split(fieldType.Tag.Get("json"), ",")[0]
			if jsonTag != "username" && jsonTag != "media_type" {
				updateExpression, expressionAttributeNames, expressionAttributeValues = addAttributeIfNotNull(updateExpression, expressionAttributeNames, expressionAttributeValues, fieldType.Name, jsonTag, field.Interface())
			}
		}
	}

	return updateExpression, expressionAttributeNames, expressionAttributeValues
}

func main() {
	lambda.Start(HandleRequest)
}
