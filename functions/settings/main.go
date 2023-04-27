package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type OptionArguments struct {
	Username            string   `json:"username" binding:"required"`
	MediaType           string   `json:"media_type"`
	ShowOnLeaderboard   *bool    `json:"show_on_leaderboard"`
	InterfaceBlurAmount *float32 `json:"interface_blur_amount"`
	MenuBlurAmount      *float32 `json:"menu_blur_amount"`
	MaxAFKTime          *float32 `json:"max_afk_time"`
	MaxBlurTime         *float32 `json:"max_blur_time"`
	MaxLoadLines        *int16   `json:"max_load_lines"`
}

type TableKey struct {
	Username  string `json:"username" binding:"required"`
	MediaType string `json:"media_type"`
}

var sess *session.Session
var svc *dynamodb.DynamoDB

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

func HandleRequest(ctx context.Context, options OptionArguments) {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)

	key := TableKey{
		Username:  options.Username,
		MediaType: options.MediaType,
	}
	tableKey, keyErr := dynamodbattribute.MarshalMap(key)

	if keyErr != nil {
		log.Fatalln("Error marshalling key:", keyErr)
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
		log.Fatalln("Error updating DynamoDB item:", updateErr)
	}

	fmt.Println(tableKey)
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

func createUpdateExpressionAttributes(optionArgs OptionArguments) (string, map[string]*string, map[string]*dynamodb.AttributeValue) {
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
