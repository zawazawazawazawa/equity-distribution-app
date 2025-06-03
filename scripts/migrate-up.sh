#!/bin/bash

# Migration実行スクリプト
# 使用方法: ./scripts/migrate-up.sh

echo "Running database migrations..."

cd backend
go run cmd/migrate/main.go up

echo "Migration completed."
