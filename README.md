# Equity Distribution App

## 開発環境のセットアップ

### 必要なツール

- Docker
- AWS CLI
- jq

### DynamoDB の操作

#### 1. LocalStack の起動

```bash
docker compose up -d
```

#### 2. テーブルの確認

```bash
# テーブル一覧
aws --endpoint-url=http://localhost:4566 dynamodb list-tables

# テーブルの中身
aws --endpoint-url=http://localhost:4566 dynamodb scan --table-name PloEquity
```

#### 3. データの更新手順

a. データのエクスポート

```bash
aws --endpoint-url=http://localhost:4566 dynamodb scan \
  --table-name PloEquity \
  --output json > export.json
```

b. エクスポートしたデータをインポート用に変換

```bash
./scripts/convert-export.sh
```

c. 変換したデータをインポート

```bash
aws --endpoint-url=http://localhost:4566 dynamodb batch-write-item \
  --request-items file://import-items.json
```

### 補足

- `export.json`と`import-items.json`は Git 管理対象外です
- データ形式のサンプルは`export.json.example`を参照してください

```bash
# テーブル一覧の確認
aws dynamodb list-tables --endpoint-url http://localhost:4566 --region us-east-1

# テーブルの中身の確認
aws dynamodb describe-table --table-name PloEquity --endpoint-url http://localhost:4566 --region us-east-1
aws dynamodb scan --table-name PloEquity --endpoint-url http://localhost:4566 --region us-east-1
```

- dynamodb のテーブルデータのエクスポート方法

  ```bash
  # JSON形式でエクスポート
  aws dynamodb scan \
    --table-name PloEquity \
    --endpoint-url http://localhost:4566 \
    --region us-east-1 \
    --output json > export.json

  # テキスト形式でエクスポート
  aws dynamodb scan \
    --table-name PloEquity \
    --endpoint-url http://localhost:4566 \
    --region us-east-1 \
    --output text > export.txt
  ```
