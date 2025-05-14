#!/bin/bash

# バッチ処理実行スクリプト
# 使用方法: ./scripts/run-batch.sh [オプション]
#
# オプション:
#   -n <数値>   : 計算回数 (デフォルト: 100)
#   -w <数値>   : ワーカー数 (デフォルト: CPUコア数)
#   -e <URL>    : DynamoDBエンドポイント (デフォルト: http://localhost:4566)
#   -r <リージョン> : AWSリージョン (デフォルト: us-east-1)
#   -l <ファイル> : ログファイル (デフォルト: stdout)
#   -d <ディレクトリ> : データディレクトリ (デフォルト: data)
#   -h          : ヘルプを表示

# デフォルト値
NUM_CALCULATIONS=100
NUM_WORKERS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
DYNAMODB_ENDPOINT="http://localhost:4566"
DYNAMODB_REGION="us-east-1"
LOG_FILE=""
DATA_DIR="data"

# コマンドライン引数の解析
while getopts "n:w:e:r:l:d:h" opt; do
  case $opt in
    n) NUM_CALCULATIONS=$OPTARG ;;
    w) NUM_WORKERS=$OPTARG ;;
    e) DYNAMODB_ENDPOINT=$OPTARG ;;
    r) DYNAMODB_REGION=$OPTARG ;;
    l) LOG_FILE=$OPTARG ;;
    d) DATA_DIR=$OPTARG ;;
    h)
      echo "使用方法: $0 [オプション]"
      echo "オプション:"
      echo "  -n <数値>   : 計算回数 (デフォルト: 100)"
      echo "  -w <数値>   : ワーカー数 (デフォルト: CPUコア数)"
      echo "  -e <URL>    : DynamoDBエンドポイント (デフォルト: http://localhost:4566)"
      echo "  -r <リージョン> : AWSリージョン (デフォルト: us-east-1)"
      echo "  -l <ファイル> : ログファイル (デフォルト: stdout)"
      echo "  -d <ディレクトリ> : データディレクトリ (デフォルト: data)"
      echo "  -h          : ヘルプを表示"
      exit 0
      ;;
    \?)
      echo "無効なオプション: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "オプション -$OPTARG には引数が必要です" >&2
      exit 1
      ;;
  esac
done

# DynamoDBが起動しているか確認
echo "DynamoDBの接続確認中..."
aws dynamodb list-tables --endpoint-url $DYNAMODB_ENDPOINT --region $DYNAMODB_REGION &>/dev/null
if [ $? -ne 0 ]; then
  echo "エラー: DynamoDBに接続できません。docker-compose upを実行してDynamoDBを起動してください。" >&2
  exit 1
fi

# バッチ処理の実行
echo "バッチ処理を開始します..."
echo "計算回数: $NUM_CALCULATIONS"
echo "ワーカー数: $NUM_WORKERS"
echo "DynamoDBエンドポイント: $DYNAMODB_ENDPOINT"
echo "AWSリージョン: $DYNAMODB_REGION"
if [ -n "$LOG_FILE" ]; then
  echo "ログファイル: $LOG_FILE"
else
  echo "ログ: 標準出力"
fi
echo "データディレクトリ: $DATA_DIR"

# コマンドライン引数の構築
CMD_ARGS=""
CMD_ARGS="$CMD_ARGS -n $NUM_CALCULATIONS"
CMD_ARGS="$CMD_ARGS -w $NUM_WORKERS"
CMD_ARGS="$CMD_ARGS -endpoint $DYNAMODB_ENDPOINT"
CMD_ARGS="$CMD_ARGS -region $DYNAMODB_REGION"
CMD_ARGS="$CMD_ARGS -data $DATA_DIR"
if [ -n "$LOG_FILE" ]; then
  CMD_ARGS="$CMD_ARGS -log $LOG_FILE"
fi

# バッチ処理の実行
cd $(dirname $0)/..
go run backend/batch/main.go $CMD_ARGS

# 終了コードの確認
if [ $? -eq 0 ]; then
  echo "バッチ処理が正常に完了しました。"
else
  echo "エラー: バッチ処理が異常終了しました。" >&2
  exit 1
fi
