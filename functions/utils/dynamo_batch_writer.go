package utils

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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
