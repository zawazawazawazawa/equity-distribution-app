package db

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Config represents DynamoDB configuration
type Config struct {
	Region   string
	Endpoint string
}

// DynamoDBAPI defines the interface for DynamoDB operations
type DynamoDBAPI interface {
	PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
}

// GetDynamoDBClient initializes and returns a DynamoDB client
func GetDynamoDBClient(config Config) *dynamodb.DynamoDB {
	// AWS設定
	awsConfig := &aws.Config{
		Region:   aws.String(config.Region),
		Endpoint: aws.String(config.Endpoint),
	}

	// 認証情報を設定（LocalStackの場合はダミーでOK）
	awsConfig.Credentials = credentials.NewStaticCredentials("test", "test", "")

	// セッションを作成
	sess := session.Must(session.NewSession(awsConfig))

	// DynamoDBクライアントを作成
	return dynamodb.New(sess)
}

// InsertDynamoDB inserts or updates an item in DynamoDB
func InsertDynamoDB(svc DynamoDBAPI, tableName string, flop string, handCombination string, equity float64) error {
	log.Printf("Inserting data - Flop: %s, HandCombination: %s, Equity: %.2f", flop, handCombination, equity)

	item := map[string]*dynamodb.AttributeValue{
		"Flop": {
			S: aws.String(flop),
		},
		"HandCombination": {
			S: aws.String(handCombination),
		},
		"Equity": {
			N: aws.String(fmt.Sprintf("%.2f", equity)),
		},
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}

	_, err := svc.PutItem(input)
	if err != nil {
		log.Printf("Error inserting data into DynamoDB: %v", err)
		return fmt.Errorf("failed to insert data into DynamoDB: %v", err)
	}
	log.Printf("Successfully inserted data into DynamoDB for %s", handCombination)
	return nil
}

// BatchQueryDynamoDB retrieves all equity calculations for a specific flop
func BatchQueryDynamoDB(svc DynamoDBAPI, tableName string, flop string) (map[string]float64, error) {
	// Query parameters for scanning items with the same flop
	log.Printf("Querying DynamoDB for flop: %s", flop)
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("Flop = :flop"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":flop": {
				S: aws.String(flop),
			},
		},
	}

	result, err := svc.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %v", err)
	}

	// Create equities map
	equities := make(map[string]float64)

	// Unmarshal each item
	for _, item := range result.Items {
		var dbItem struct {
			HandCombination string  `json:"HandCombination"`
			Equity          float64 `json:"Equity"`
		}
		err = dynamodbattribute.UnmarshalMap(item, &dbItem)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal DynamoDB item: %v", err)
		}
		equities[dbItem.HandCombination] = dbItem.Equity
	}

	return equities, nil
}
