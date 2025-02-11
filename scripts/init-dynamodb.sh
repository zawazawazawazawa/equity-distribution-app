#!/bin/bash

echo "Waiting for LocalStack to be ready..."
sleep 5

echo "Creating DynamoDB table..."
aws --endpoint-url=http://localhost:4566 dynamodb create-table \
    --table-name PloEquity \
    --attribute-definitions \
        AttributeName=Flop,AttributeType=S \
        AttributeName=HandCombination,AttributeType=S \
    --key-schema \
        AttributeName=Flop,KeyType=HASH \
        AttributeName=HandCombination,KeyType=RANGE \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

echo "Table creation completed"

echo "Importing initial data..."
# export.jsonから直接データをインポート
aws --endpoint-url=http://localhost:4566 dynamodb batch-write-item \
    --request-items file:///docker-entrypoint-initaws.d/import-items.json

echo "Data import completed"
