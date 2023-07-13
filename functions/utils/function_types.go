package utils

import "github.com/aws/aws-sdk-go/service/dynamodb"

type BatchwriteArgs struct {
	TableName     string                   `json:"table_name"`
	WriteRequests []*dynamodb.WriteRequest `json:"write_requests"`
	MaxBatchSize  int                      `json:"max_batch_size" default:"25"`
}

type IntermediateEntryItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
	UserMediaEntry
}

type IntermediateStatItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
	UserMediaStat
}
