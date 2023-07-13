package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const AWSMaxBatchSize = 25

func ZeroPadInt64(number int64) string {
	return fmt.Sprintf("%0*d", strconv.IntSize/4, number)
}

func AddAttributeIfNotNull(updateExpression string, expressionAttributeNames map[string]*string, expressionAttributeValues map[string]*dynamodb.AttributeValue, attributeName, jsonAttributeName string, value interface{}) (string, map[string]*string, map[string]*dynamodb.AttributeValue) {
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
				updateExpression, expressionAttributeNames, expressionAttributeValues = AddAttributeIfNotNull(updateExpression, expressionAttributeNames, expressionAttributeValues, fieldType.Name, jsonTag, field.Interface())
			}
		}
	}

	return updateExpression, expressionAttributeNames, expressionAttributeValues
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

func BatchWrite(svc *dynamodb.DynamoDB, tableName string, items []*dynamodb.WriteRequest) []*dynamodb.WriteRequest {
	unprocessedWrites := []*dynamodb.WriteRequest{}

	if len(items) > AWSMaxBatchSize {
		unprocessedWrites = append(unprocessedWrites, items[AWSMaxBatchSize:]...)
	}

	output, err := svc.BatchWriteItem(&dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			tableName: items[:min(AWSMaxBatchSize, len(items))],
		},
	})
	unprocessedWrites = append(unprocessedWrites, output.UnprocessedItems[tableName]...)

	if err != nil {
		itemsArray := zerolog.Arr()

		for _, item := range items {
			itemsArray.Interface(item.PutRequest.Item)
		}

		log.Error().Err(err).Str("table_name", tableName).Array("items", itemsArray).Msg("Dynamodb batch write failed")
	} else {
		log.Info().Str("table_name", tableName).Msg("Dynamodb batch write succeeded")
	}

	return unprocessedWrites
}

func DistributedBatchWrites(svc *dynamodb.DynamoDB, batchwriteArgs *BatchwriteArgs) *BatchwriteArgs {
	if batchwriteArgs.MaxBatchSize < 1 || batchwriteArgs.MaxBatchSize > AWSMaxBatchSize {
		log.Info().Str("table_name", batchwriteArgs.TableName).Int("max_batch_size", batchwriteArgs.MaxBatchSize).Msg("Batch writes attempted with invalid max batch size")
		batchwriteArgs.MaxBatchSize = AWSMaxBatchSize
	}

	var waitGroup sync.WaitGroup
	channel := make(chan []*dynamodb.WriteRequest)

	// Process each batch in separate threads
	for start := 0; start < len(batchwriteArgs.WriteRequests); start += batchwriteArgs.MaxBatchSize {
		waitGroup.Add(1)

		go func(start int) {
			defer waitGroup.Done()

			end := min(start+batchwriteArgs.MaxBatchSize, len(batchwriteArgs.WriteRequests))
			channel <- BatchWrite(svc, batchwriteArgs.TableName, batchwriteArgs.WriteRequests[start:end])
		}(start)
	}

	// Waiting thread to close the channel
	// Once all batches have been tried
	go func() {
		waitGroup.Wait()
		close(channel)
	}()

	// Read from the channel concatenating unprocessed writes
	unprocessedWrites := []*dynamodb.WriteRequest{}
	for unprocessedWriteBatch := range channel {
		unprocessedWrites = append(unprocessedWrites, unprocessedWriteBatch...)
	}

	return &BatchwriteArgs{
		TableName:     batchwriteArgs.TableName,
		WriteRequests: unprocessedWrites,
		MaxBatchSize:  batchwriteArgs.MaxBatchSize,
	}
}
