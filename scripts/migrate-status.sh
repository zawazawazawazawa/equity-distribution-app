#!/bin/bash

# Migration ステータス確認スクリプト
# 使用方法: ./scripts/migrate-status.sh

echo "Checking migration status..."

cd backend
go run cmd/migrate/main.go status
