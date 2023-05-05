package dynamo_types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const AWSMaxBatchSize = 25

type CompositeKey struct {
	PK interface{} `json:"pk" binding:"required"`
	SK interface{} `json:"sk"`
}

type UserSettingsKey struct {
	Username  string `json:"username" binding:"required"`
	MediaType string `json:"media_type"`
}

type UserSettings struct {
	Key                 UserSettingsKey `json:"key" binding:"required"`
	ShowOnLeaderboard   *bool           `json:"show_on_leaderboard"`
	InterfaceBlurAmount *float32        `json:"interface_blur_amount"`
	MenuBlurAmount      *float32        `json:"menu_blur_amount"`
	MaxAFKTime          *int16          `json:"max_afk_time"`
	MaxBlurTime         *int16          `json:"max_blur_time"`
	MaxLoadLines        *int16          `json:"max_load_lines"`
}

type UserMediaKey struct {
	Username        string `json:"username" binding:"required"`
	MediaType       string `json:"media_type" binding:"required"`
	MediaIdentifier string `json:"media_identifier"`
}

type MediaStat struct {
	TimeRead  int64 `json:"time_read" binding:"required"`
	CharsRead int64 `json:"chars_read" binding:"required"`
	LinesRead int64 `json:"lines_read"`
}

type UserMediaEntry struct {
	Key         UserMediaKey `json:"key" binding:"required"`
	DisplayName string       `json:"display_name"`
	Series      string       `json:"series"`
	LastUpdate  int64        `json:"last_update"`
}

type UserMediaStat struct {
	Key        UserMediaKey `json:"key" binding:"required"`
	Date       *int64       `json:"date"`
	Stats      MediaStat    `json:"stats"`
	LastUpdate int64        `json:"last_update"`
	Pause      bool         `json:"pause"`
}

type LeaderboardKey struct {
	Username   string `json:"username" binding:"required"`
	TimePeriod string `json:"time_period" binding:"required"`
	MediaType  string `json:"media_type" binding:"required"`
}

type LeaderboardEntry struct {
	Key        LeaderboardKey `json:"key" binding:"required"`
	MediaNames string         `json:"media_names"`
	TimeRead   int64          `json:"time_read"`
	CharsRead  int64          `json:"chars_read"`
}

type BatchwriteArgs struct {
	TableName     string                   `json:"table_name"`
	WriteRequests []*dynamodb.WriteRequest `json:"write_requests"`
	MaxBatchSize  int                      `json:"max_batch_size" default:"25"`
}

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

func PutItemRequest(svc *dynamodb.DynamoDB, tableKey map[string]*dynamodb.AttributeValue, itemData interface{}) (*dynamodb.WriteRequest, error) {
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
			tableName: items[:AWSMaxBatchSize],
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

			channel <- BatchWrite(svc, batchwriteArgs.TableName, batchwriteArgs.WriteRequests[start:start+batchwriteArgs.MaxBatchSize])
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
