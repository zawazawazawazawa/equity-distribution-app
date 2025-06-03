# Equity Distribution App

## 開発環境のセットアップ

### 必要なツール

- Docker
- AWS CLI
- jq
- Go 1.23+

### PostgreSQL の操作

#### 1. PostgreSQL の起動

```bash
docker compose up -d postgres
```

#### 2. データベースマイグレーション

```bash
# マイグレーション適用
./scripts/migrate-up.sh

# ロールバック
./scripts/migrate-down.sh

# ステータス確認
./scripts/migrate-status.sh
```

#### 3. データベース接続

```bash
# PostgreSQLに直接接続
docker exec -it postgres_plo psql -U postgres -d plo_equity
```

詳細なマイグレーション管理については、[backend/migrations/README.md](backend/migrations/README.md) を参照してください。

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
