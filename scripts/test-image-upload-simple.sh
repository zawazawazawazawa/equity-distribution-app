#!/bin/bash

# 画像アップロードテスト用シンプルスクリプト

# 現在のディレクトリを取得
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# バックエンドディレクトリに移動
cd "$PROJECT_ROOT/backend" || exit 1

# 引数の解析
DATE=""
SCENARIO=""
HERO_HAND=""
FLOP=""
LOG_FILE=""

# ヘルプメッセージ
show_help() {
  echo "使用方法: $0 [オプション]"
  echo ""
  echo "オプション:"
  echo "  -h, --help                このヘルプメッセージを表示"
  echo "  -d, --date DATE           日付を指定 (YYYY-MM-DD形式、デフォルト: 今日)"
  echo "  -s, --scenario SCENARIO   シナリオ名を指定 (デフォルト: 'PLO5 SRP UTG vs BB')"
  echo "  -H, --hand HAND           ヒーローハンドを指定 (デフォルト: 'AsKsQsJsTs')"
  echo "  -f, --flop FLOP           フロップを指定 (例: '2c3d4h')"
  echo "  -l, --log FILE            ログファイルを指定"
  echo ""
  echo "例:"
  echo "  # 5カードPLOのテスト"
  echo "  $0 --hand AsKsQsJsTs --flop 2h3s4c"
  echo "  # 4カードPLOのテスト"
  echo "  $0 --hand AsKsQsJs --flop TdJdQd --scenario \"SRP UTG vs BB\""
  exit 0
}

# 引数の解析
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      show_help
      ;;
    -d|--date)
      DATE="$2"
      shift 2
      ;;
    -s|--scenario)
      SCENARIO="$2"
      shift 2
      ;;
    -H|--hand)
      HERO_HAND="$2"
      shift 2
      ;;
    -f|--flop)
      FLOP="$2"
      shift 2
      ;;
    -l|--log)
      LOG_FILE="$2"
      shift 2
      ;;
    *)
      echo "エラー: 不明なオプション '$1'"
      show_help
      ;;
  esac
done

# コマンドライン引数の構築
CMD_ARGS=""

if [ -n "$DATE" ]; then
  CMD_ARGS="$CMD_ARGS -date=$DATE"
fi

if [ -n "$SCENARIO" ]; then
  CMD_ARGS="$CMD_ARGS -scenario=\"$SCENARIO\""
fi

if [ -n "$HERO_HAND" ]; then
  CMD_ARGS="$CMD_ARGS -hand=$HERO_HAND"
fi

if [ -n "$FLOP" ]; then
  CMD_ARGS="$CMD_ARGS -flop=$FLOP"
fi

if [ -n "$LOG_FILE" ]; then
  CMD_ARGS="$CMD_ARGS -log=$LOG_FILE"
fi

# R2設定の環境変数が設定されているか確認
if [ -z "$R2_ENDPOINT" ] || [ -z "$R2_ACCESS_KEY" ] || [ -z "$R2_SECRET_KEY" ] || [ -z "$R2_BUCKET" ]; then
  echo "警告: R2設定の環境変数が設定されていません。画像生成のみが実行されます。"
  echo "R2へのアップロードをテストするには、以下の環境変数を設定してください:"
  echo "  export R2_ENDPOINT=your_endpoint"
  echo "  export R2_ACCESS_KEY=your_access_key"
  echo "  export R2_SECRET_KEY=your_secret_key"
  echo "  export R2_BUCKET=your_bucket_name"
  echo ""
fi

# テストプログラムの実行
echo "画像アップロードテストを実行しています..."
echo "実行コマンド: go run cmd/test-image-upload/main.go $CMD_ARGS"
eval "go run cmd/test-image-upload/main.go $CMD_ARGS"

# 終了コードの確認
if [ $? -eq 0 ]; then
  echo ""
  echo "テストが正常に完了しました"
else
  echo ""
  echo "エラー: テストの実行中にエラーが発生しました"
  exit 1
fi