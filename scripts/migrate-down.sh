#!/bin/bash

# Migration ロールバックスクリプト
# 使用方法: ./scripts/migrate-down.sh

echo "Rolling back last migration..."

cd backend
go run cmd/migrate/main.go down

echo "Rollback completed."
