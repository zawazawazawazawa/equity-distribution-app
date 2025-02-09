#!/bin/bash

echo "Waiting for LocalStack to be ready..."
sleep 10

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
