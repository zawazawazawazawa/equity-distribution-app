# PostgreSQL 統合計画

## 概要

このドキュメントでは、既存の DynamoDB と並行して PostgreSQL をプロジェクトに追加するための計画を説明します。バッチ処理の結果を PostgreSQL に保存することが目的です。

## 1. docker-compose.yml の修正

以下のように docker-compose.yml ファイルを修正して、PostgreSQL サービスを追加します：

```yaml
version: "3.8"

services:
  localstack:
    image: localstack/localstack:latest
    container_name: localstack_plo
    ports:
      - "4566:4566" # LocalStack Gateway
      - "4510-4559:4510-4559" # external services port range
    environment:
      - SERVICES=dynamodb
      - DEBUG=1
      - DATA_DIR=/tmp/localstack/data
      - DOCKER_HOST=unix:///var/run/docker.sock
      - AWS_DEFAULT_REGION=us-east-1
      - AWS_ACCESS_KEY_ID=dummy
      - AWS_SECRET_ACCESS_KEY=dummy
    volumes:
      - "${LOCALSTACK_VOLUME_DIR:-./volume}:/var/lib/localstack"
      - "/var/run/docker.sock:/var/run/docker.sock"
      - ./scripts:/docker-entrypoint-initaws.d # スクリプトディレクトリをマウント
      - ./export.json:/docker-entrypoint-initaws.d/export.json
      - ./import-items.json:/docker-entrypoint-initaws.d/import-items.json
    entrypoint: ""
    command: |
      /bin/sh -c '
        /usr/local/bin/docker-entrypoint.sh &
        echo "Waiting for LocalStack to be ready..."
        while ! awslocal dynamodb list-tables 2>/dev/null; do
          echo "Waiting for LocalStack..."
          sleep 2
        done
        echo "LocalStack is ready! Running initialization scripts..."
        sh /docker-entrypoint-initaws.d/init-dynamodb.sh
        tail -f /dev/null
      '

  # PostgreSQLサービスを追加
  postgres:
    image: postgres:latest
    container_name: postgres_plo
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=plo_equity
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./postgres/init:/docker-entrypoint-initdb.d

volumes:
  localstack-vol:
    driver: local
  postgres-data:
    driver: local
```

## 2. PostgreSQL 初期化スクリプトの作成

PostgreSQL の初期化スクリプトを作成するためのディレクトリとファイルを作成します：

```
postgres/
└── init/
    └── init.sql
```

init.sql には、バッチ処理の結果を保存するためのテーブルを作成する SQL スクリプトを記述します：

```sql
-- バッチ処理の結果を保存するテーブル
CREATE TABLE IF NOT EXISTS daily_quiz_results (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    scenario VARCHAR(255) NOT NULL,
    hero_hand VARCHAR(255) NOT NULL,
    flop VARCHAR(255) NOT NULL,
    result TEXT,
    average_equity DECIMAL(5,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- インデックスの作成
CREATE INDEX IF NOT EXISTS idx_daily_quiz_results_date ON daily_quiz_results(date);
CREATE INDEX IF NOT EXISTS idx_daily_quiz_results_scenario ON daily_quiz_results(scenario);
CREATE INDEX IF NOT EXISTS idx_daily_quiz_results_hero_hand ON daily_quiz_results(hero_hand);
CREATE INDEX IF NOT EXISTS idx_daily_quiz_results_flop ON daily_quiz_results(flop);
```

## 3. 実装手順

1. docker-compose.yml ファイルを修正
2. PostgreSQL 初期化スクリプトのディレクトリとファイルを作成
   ```bash
   mkdir -p postgres/init
   touch postgres/init/init.sql
   ```
3. init.sql ファイルに上記の SQL スクリプトを記述
4. 修正した docker-compose.yml でコンテナを起動
   ```bash
   docker-compose down
   docker-compose up -d
   ```
5. PostgreSQL が正常に動作することを確認
   ```bash
   docker exec -it postgres_plo psql -U postgres -d plo_equity -c "SELECT * FROM daily_quiz_results LIMIT 5;"
   ```

## 4. 将来的な対応（バックエンドコードの修正）

バックエンドコードに PostgreSQL との接続機能を追加する必要があります：

1. PostgreSQL クライアントライブラリの追加
2. データベース接続設定の追加
3. PostgreSQL へのデータ保存・取得機能の実装

これらの対応は今回の範囲外ですが、将来的に実装する必要があります。
