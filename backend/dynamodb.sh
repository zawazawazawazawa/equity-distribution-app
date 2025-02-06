#!/bin/bash
aws dynamodb create-table \
    --table-name PloEquity \
    --attribute-definitions \
        AttributeName=Flop,AttributeType=S \
        AttributeName=HandCombination,AttributeType=S \
    --key-schema \
        AttributeName=Flop,KeyType=HASH \
        AttributeName=HandCombination,KeyType=RANGE \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:4566
