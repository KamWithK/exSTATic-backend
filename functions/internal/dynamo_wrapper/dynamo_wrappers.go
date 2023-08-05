package dynamo_wrapper

import (
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog/log"
)

const AWSMaxBatchSize = 25

type CompositeKey struct {
	PK interface{} `json:"pk" binding:"required"`
	SK interface{} `json:"sk"`
}

type BatchwriteArgs struct {
	TableName     string                   `json:"table_name"`
	WriteRequests []*dynamodb.WriteRequest `json:"write_requests"`
	MaxBatchSize  int                      `json:"max_batch_size" default:"25"`
}

func GetCompositeKey(pk interface{}, sk interface{}) (map[string]*dynamodb.AttributeValue, error) {
	var compositeKey = CompositeKey{
		PK: pk,
		SK: sk,
	}

	tableKey, keyErr := dynamodbattribute.MarshalMap(compositeKey)
	if keyErr != nil {
		log.Error().Err(keyErr).Interface("pk", pk).Interface("sk", sk).Msg("Could not marshal dynamodb key")
		return nil, keyErr
	}

	return tableKey, nil
}

func RemoveNullAttributes(updateExpression *string, expressionAttributeNames map[string]*string, expressionAttributeValues map[string]*dynamodb.AttributeValue, attributeName, jsonAttributeName string, value interface{}) {
	if value != nil {
		if len(expressionAttributeNames) > 0 {
			*updateExpression += ","
		}
		*updateExpression += " #" + attributeName + " = :" + attributeName
		expressionAttributeNames["#"+attributeName] = aws.String(jsonAttributeName)
		value, _ := dynamodbattribute.Marshal(value)
		expressionAttributeValues[":"+attributeName] = value
	}
}

func CreateUpdateExpressionAttributes(optionArgs interface{}) (string, map[string]*string, map[string]*dynamodb.AttributeValue) {
	updateExpression := "SET"
	expressionAttributeNames := map[string]*string{}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{}

	valueOfOptionArgs := reflect.ValueOf(optionArgs)
	typeOfOptionArgs := valueOfOptionArgs.Type()

	for i := 0; i < valueOfOptionArgs.NumField(); i++ {
		field := valueOfOptionArgs.Field(i)
		fieldType := typeOfOptionArgs.Field(i)

		if field.Kind() != reflect.Invalid && !field.IsZero() {
			if fieldType.Name != "Key" {
				jsonTag := strings.Split(fieldType.Tag.Get("json"), ",")[0]
				RemoveNullAttributes(&updateExpression, expressionAttributeNames, expressionAttributeValues, fieldType.Name, jsonTag, field.Interface())
			}
		}
	}

	return updateExpression, expressionAttributeNames, expressionAttributeValues
}

func CombineAttributes(firstAttributes map[string]*dynamodb.AttributeValue, secondAttributes map[string]*dynamodb.AttributeValue) map[string]*dynamodb.AttributeValue {
	combinedAttributes := map[string]*dynamodb.AttributeValue{}

	for k, v := range firstAttributes {
		combinedAttributes[k] = v
	}
	for k, v := range secondAttributes {
		combinedAttributes[k] = v
	}

	return combinedAttributes
}

func UpdateItem(svc *dynamodb.DynamoDB, tableName string, tableKey map[string]*dynamodb.AttributeValue, tableData interface{}) (*dynamodb.UpdateItemOutput, error) {
	// Get dynamodb query information
	updateExpression, expressionAttributeNames, expressionAttributeValues := CreateUpdateExpressionAttributes(tableData)

	// Put item
	updateItem, updateErr := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String(tableName),
		Key:                       tableKey,
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	})
	if updateErr != nil {
		log.Error().Err(updateErr).Str("table_name", tableName).Interface("table_key", tableKey).Interface("item", tableData).Msg("Dynamodb update item errored")
		return nil, updateErr
	}

	return updateItem, nil
}

func PutItem(svc *dynamodb.DynamoDB, tableName string, tableKey map[string]*dynamodb.AttributeValue, itemData interface{}) (*dynamodb.PutItemOutput, error) {
	// Convert item data to DynamoDB attribute values
	itemAttributes, err := dynamodbattribute.MarshalMap(itemData)
	if err != nil {
		log.Error().Err(err).Interface("item", itemData).Msg("Could not marshal dynamodb item")
		return nil, err
	}

	delete(itemAttributes, "key")

	// Put the item
	putItem, putErr := svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      CombineAttributes(tableKey, itemAttributes),
	})
	if putErr != nil {
		log.Error().Err(putErr).Str("table_name", tableName).Interface("table_key", tableKey).Interface("item", itemData).Msg("Dynamodb put item errored")
		return nil, putErr
	}

	return putItem, nil
}

func PutItemRequest(tableKey map[string]*dynamodb.AttributeValue, itemData interface{}) (*dynamodb.WriteRequest, error) {
	// Convert item data to DynamoDB attribute values
	itemAttributes, err := dynamodbattribute.MarshalMap(itemData)
	if err != nil {
		log.Error().Err(err).Interface("item", itemData).Msg("Could not marshal dynamodb item")
		return nil, err
	}

	delete(itemAttributes, "key")

	// Put the item
	return &dynamodb.WriteRequest{
		PutRequest: &dynamodb.PutRequest{
			Item: CombineAttributes(tableKey, itemAttributes),
		},
	}, nil
}

func PutRawRequest(pk string, sk string, userMedia interface{}) *dynamodb.WriteRequest {
	tableKey, keyErr := GetCompositeKey(pk, sk)
	if keyErr != nil {
		return nil
	}

	writeRequest, writeErr := PutItemRequest(tableKey, userMedia)

	if writeErr != nil {
		return nil
	}

	return writeRequest
}
