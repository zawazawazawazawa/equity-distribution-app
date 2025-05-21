#!/bin/bash

# test-image-generator.shスクリプト
# 使用方法: ./scripts/test-image-generator.sh

# スクリプトのあるディレクトリに移動
cd $(dirname $0)/..

# ビルドと実行
echo "画像生成テストを開始します..."
(cd backend && go run cmd/test-image-generator/main.go)

# 終了ステータスの確認
if [ $? -eq 0 ]; then
  echo "テストが正常に完了しました。"
  echo "生成された画像は ./backend/images/daily-quiz/ ディレクトリにあります。"
else
  echo "テストが失敗しました。"
  exit 1
fi
