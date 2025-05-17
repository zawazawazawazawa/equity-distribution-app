#!/bin/bash

# バッチ処理実行スクリプト
# 使用方法: ./scripts/run-batch.sh [オプション]
#
# オプション:
#   -l <ファイル>    : ログファイル (デフォルト: stdout)
#   -d <ディレクトリ> : データディレクトリ (デフォルト: data)
#   -D <YYYY-MM-DD> : 開始日付 (デフォルト: 翌日)
#   -N <日数>       : 生成する日数 (デフォルト: 7日間、開始日から1週間)
#   -H <ホスト>      : PostgreSQLホスト (デフォルト: localhost)
#   -p <ポート>      : PostgreSQLポート (デフォルト: 5432)
#   -u <ユーザー>    : PostgreSQLユーザー (デフォルト: postgres)
#   -P <パスワード>  : PostgreSQLパスワード (デフォルト: postgres)
#   -n <DB名>       : PostgreSQLデータベース名 (デフォルト: plo_equity)
#   -h              : ヘルプを表示

# 日付生成関数
generate_date() {
  local base_date="$1"
  local days_offset="$2"
  
  # OSを検出して適切なdate書式を使用
  if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    date -v+${days_offset}d -j -f "%Y-%m-%d" "$base_date" "+%Y-%m-%d"
  else
    # Linux
    date -d "$base_date + $days_offset days" "+%Y-%m-%d"
  fi
}

# デフォルト値（環境変数から取得、設定されていない場合はデフォルト値を使用）
LOG_FILE=""
DATA_DIR="data"
DATE=""
DAYS_COUNT=7  # デフォルトは7日間
POSTGRES_HOST=${POSTGRES_HOST:-"localhost"}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_USER=${POSTGRES_USER:-"postgres"}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-"postgres"}
POSTGRES_DBNAME=${POSTGRES_DBNAME:-"plo_equity"}

# コマンドライン引数の解析
while getopts "l:d:D:N:H:p:u:P:n:h" opt; do
  case $opt in
    l) LOG_FILE=$OPTARG ;;
    d) DATA_DIR=$OPTARG ;;
    D) DATE=$OPTARG ;;
    N) DAYS_COUNT=$OPTARG ;;
    H) POSTGRES_HOST=$OPTARG ;;
    p) POSTGRES_PORT=$OPTARG ;;
    u) POSTGRES_USER=$OPTARG ;;
    P) POSTGRES_PASSWORD=$OPTARG ;;
    n) POSTGRES_DBNAME=$OPTARG ;;
    h)
      echo "使用方法: $0 [オプション]"
      echo "オプション:"
      echo "  -l <ファイル>    : ログファイル (デフォルト: stdout)"
      echo "  -d <ディレクトリ> : データディレクトリ (デフォルト: data)"
      echo "  -D <YYYY-MM-DD> : 開始日付 (デフォルト: 翌日)"
      echo "  -N <日数>       : 生成する日数 (デフォルト: 7日間、開始日から1週間)"
      echo "  -H <ホスト>      : PostgreSQLホスト (デフォルト: localhost)"
      echo "  -p <ポート>      : PostgreSQLポート (デフォルト: 5432)"
      echo "  -u <ユーザー>    : PostgreSQLユーザー (デフォルト: postgres)"
      echo "  -P <パスワード>  : PostgreSQLパスワード (デフォルト: postgres)"
      echo "  -n <DB名>       : PostgreSQLデータベース名 (デフォルト: plo_equity)"
      echo "  -h              : ヘルプを表示"
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

# 開始日付の設定
START_DATE=$DATE
if [ -z "$START_DATE" ]; then
  # デフォルトは翌日
  if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    START_DATE=$(date -v+1d "+%Y-%m-%d")
  else
    # Linux
    START_DATE=$(date -d "tomorrow" "+%Y-%m-%d")
  fi
fi

# バッチ処理の実行
echo "バッチ処理を開始します..."
echo "データディレクトリ: $DATA_DIR"
echo "開始日付: $START_DATE"
echo "処理日数: $DAYS_COUNT 日間"
if [ -n "$LOG_FILE" ]; then
  echo "ログファイル: $LOG_FILE"
else
  echo "ログ: 標準出力"
fi

# ディレクトリ移動（ループの前に一度だけ実行）
cd $(dirname $0)/..

# 各日付に対してバッチ処理を実行
for ((i=0; i<DAYS_COUNT; i++)); do
  # 日付を生成
  CURRENT_DATE=$(generate_date "$START_DATE" $i)
  
  echo "===== 日付: $CURRENT_DATE の処理を開始します ====="
  
  # コマンドライン引数の構築
  CMD_ARGS=""
  if [ -n "$LOG_FILE" ]; then
    CMD_ARGS="$CMD_ARGS -log $LOG_FILE"
  fi
  CMD_ARGS="$CMD_ARGS -data $DATA_DIR"
  CMD_ARGS="$CMD_ARGS -date $CURRENT_DATE"
  CMD_ARGS="$CMD_ARGS -pg-host $POSTGRES_HOST"
  CMD_ARGS="$CMD_ARGS -pg-port $POSTGRES_PORT"
  CMD_ARGS="$CMD_ARGS -pg-user $POSTGRES_USER"
  CMD_ARGS="$CMD_ARGS -pg-password $POSTGRES_PASSWORD"
  CMD_ARGS="$CMD_ARGS -pg-dbname $POSTGRES_DBNAME"
  
  # バッチ処理の実行
  echo "$(date '+%Y-%m-%d %H:%M:%S') - 日付: $CURRENT_DATE の処理を開始します"
  # backendディレクトリに移動してからバッチ処理を実行
  if (cd backend && go run batch/main.go $CMD_ARGS); then
    echo "$(date '+%Y-%m-%d %H:%M:%S') - 日付: $CURRENT_DATE の処理が正常に完了しました。"
  else
    echo "$(date '+%Y-%m-%d %H:%M:%S') - エラー: 日付: $CURRENT_DATE の処理が失敗しました。次の日付に進みます。" >&2
  fi
  
  echo "===== 日付: $CURRENT_DATE の処理が完了しました ====="
  echo ""
done

# 終了メッセージ
echo "すべての日付（$DAYS_COUNT 日間）の処理が完了しました。"
