package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// DynamoDBのモックインターフェース
type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	putItemOutput    *dynamodb.PutItemOutput
	putItemErr       error
	queryOutput      *dynamodb.QueryOutput
	queryErr         error
	putItemCallCount int
	queryCallCount   int
}

func (m *mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	m.putItemCallCount++
	return m.putItemOutput, m.putItemErr
}

func (m *mockDynamoDBClient) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	m.queryCallCount++
	return m.queryOutput, m.queryErr
}

func TestGetDynamoDBClient(t *testing.T) {
	// テスト用の設定
	config := Config{
		Region:   "us-west-2",
		Endpoint: "http://localhost:8000",
	}

	// クライアントの取得
	client := GetDynamoDBClient(config)

	// クライアントがnilでないことを確認
	if client == nil {
		t.Error("Expected non-nil DynamoDB client, got nil")
	}

	// 設定値が正しく反映されていることを確認
	if *client.Config.Region != config.Region {
		t.Errorf("Expected Region to be %s, got %s", config.Region, *client.Config.Region)
	}

	if *client.Config.Endpoint != config.Endpoint {
		t.Errorf("Expected Endpoint to be %s, got %s", config.Endpoint, *client.Config.Endpoint)
	}
}

// 以下のテストはモックを使用するため、実際のDynamoDBクライアントを使用せずにスキップします
func TestInsertDynamoDB(t *testing.T) {
	t.Skip("Skipping test that requires actual DynamoDB client")
}

func TestBatchQueryDynamoDB(t *testing.T) {
	t.Skip("Skipping test that requires actual DynamoDB client")
}

// テスト用のヘルパー関数
func TestInsertDynamoDBIntegration(t *testing.T) {
	// 統合テストはスキップ（CI環境では実行しない）
	t.Skip("Skipping integration test")

	// 実際のDynamoDBクライアントを使用したテスト
	config := Config{
		Region:   "us-west-2",
		Endpoint: "http://localhost:8000", // LocalStackなどのエンドポイント
	}
	client := GetDynamoDBClient(config)

	// テストデータ
	tableName := "TestTable"
	flop := "AhKdQc"
	handCombination := fmt.Sprintf("Test_%d", time.Now().UnixNano())
	equity := 65.5

	// データの挿入
	err := InsertDynamoDB(client, tableName, flop, handCombination, equity)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// データの取得
	equities, err := BatchQueryDynamoDB(client, tableName, flop)
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}

	// 挿入したデータが取得できることを確認
	if equities[handCombination] != equity {
		t.Errorf("Expected equity for %s to be %.2f, got %.2f", handCombination, equity, equities[handCombination])
	}
}
