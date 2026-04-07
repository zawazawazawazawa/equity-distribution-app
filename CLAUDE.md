# CLAUDE.md

このファイルは、Claude Code（claude.ai/code）がこのリポジトリのコードを扱う際のガイダンスを提供します。

## プロジェクト概要

これはポーカーゲーム用のエクイティ分布計算アプリケーションで、特にPLO（ポットリミットオマハ）と7カードスタッドのバリアントに焦点を当てています。全数計算、モンテカルロシミュレーション、適応サンプリングなど、様々なアルゴリズムを使用してエクイティ分布を計算します。

## 技術スタック

- **バックエンド**: Go 1.23+ と Gin Webフレームワーク
- **データベース**: PostgreSQL（メインストレージ）とDynamoDB（LocalStack経由でローカル開発用）
- **インフラ**: ローカル開発用のDocker Compose
- **画像ストレージ**: AWS S3/R2互換ストレージ

## よく使う開発コマンド

### バックエンド開発

```bash
# APIサーバーの起動
cd backend && go run cmd/api/main.go

# 全テストの実行
cd backend && go test ./...

# 特定パッケージのテスト実行
cd backend && go test ./pkg/poker/...

# カバレッジ付きテスト実行
cd backend && go test -cover ./...

# 特定のテストのみ実行
cd backend && go test -run TestFunctionName ./pkg/poker
```

### データベース操作

```bash
# PostgreSQLの起動
docker compose up -d postgres

# 全マイグレーションの適用
./scripts/migrate-up.sh
# または直接: cd backend && go run cmd/migrate/main.go up

# 最後のマイグレーションをロールバック
./scripts/migrate-down.sh
# または直接: cd backend && go run cmd/migrate/main.go down

# マイグレーションステータスの確認
./scripts/migrate-status.sh
# または直接: cd backend && go run cmd/migrate/main.go status

# PostgreSQLへの接続
docker exec -it postgres_plo psql -U postgres -d plo_equity
```

### バッチ処理

```bash
# デイリークイズ生成のバッチ処理実行
./scripts/run-batch.sh [オプション]

# よく使うオプション:
# -D <YYYY-MM-DD>  : 開始日付
# -N <日数>        : 処理する日数
# -M               : モンテカルロモードを有効化
# -A               : 適応サンプリングを有効化
```

### 画像生成テスト

```bash
# 画像生成のテスト
./scripts/test-image-generator.sh

# 画像アップロードのテスト
./scripts/test-image-upload.sh
```

## アーキテクチャ概要

### パッケージ構造

バックエンドは関心の分離を明確にしたクリーンアーキテクチャパターンに従っています：

- **cmd/**: 各実行可能ファイルのエントリーポイント
  - `api/`: メインAPIサーバー
  - `migrate/`: データベースマイグレーションツール
  - `test-image-generator/`: 画像生成テストユーティリティ
  - `test-image-upload/`: S3/R2アップロードテストユーティリティ

- **pkg/**: コアビジネスロジックと共有パッケージ
  - `api/`: HTTPハンドラーとルーティング
    - PLOとスタッドゲーム両方のエクイティ計算を処理
    - 精度モード実装: fast、normal、accurate、adaptive
  - `poker/`: コアポーカーロジック
    - エクイティ計算アルゴリズム（全数計算、モンテカルロ、適応サンプリング）
    - ハイとローゲーム両方のハンド評価
    - 勝者判定関数
    - 特殊なスタッドゲーム実装（Razz、Stud High、Stud Hi-Lo 8）
  - `db/`: データベース接続と操作
  - `fileio/`: プリセットレンジ用のCSVファイル処理
  - `image/`: デイリークイズ画像生成
  - `storage/`: R2/S3ストレージ統合
  - `models/`: 共有データ型

- **data/**: 異なるゲーム形式のプリセットレンジデータ
  - 6人制100bbゲーム用のPLO4とPLO5レンジ
  - ポジションとベッティングパターンで整理されたCSVファイル

- **migrations/**: PostgreSQLスキーママイグレーション
  - アップ/ダウンマイグレーション用のバージョン付きSQLファイル
  - デイリークイズ結果とゲームタイプ追跡のサポート

### APIエンドポイント

APIはバージョン管理されており、以下の主要エンドポイントを提供します：

- `GET /health`: ヘルスチェック
- `POST /api/v1/equity`: PLOエクイティ計算
- `POST /api/v1/stud/equity`: スタッドゲームエクイティ計算（単一対戦相手）
- `POST /api/v1/stud/range-equity`: 複数対戦相手に対するスタッドエクイティ計算

### 計算モード

アプリケーションは複数の計算精度レベルをサポートしています：
- **Fast**: クイック推定用の限定的な反復
- **Normal**: スピードと精度のバランス
- **Accurate**: 正確な結果のための高反復回数
- **Adaptive**: 十分な精度に達したら停止する収束ベースの計算

### データベーススキーマ

golang-migrate経由で管理されるマイグレーションを使用したPostgreSQL。主要テーブル：
- `daily_quiz_results`: ゲームタイプサポート付きクイズ結果を保存
- サポートゲームタイプ: 'plo4'、'plo5'、'razz'、'7card_stud_high'、'7card_stud_highlow8'

## 環境変数

データベース接続（デフォルト値表示）：
- `POSTGRES_HOST`: localhost
- `POSTGRES_PORT`: 5432
- `POSTGRES_USER`: postgres
- `POSTGRES_PASSWORD`: postgres
- `POSTGRES_DBNAME`: plo_equity

## テスト戦略

コードベースには以下の包括的なユニットテストが含まれています：
- ポーカーハンド評価とエクイティ計算
- RazzとHi-Loゲーム用のローハンド評価
- APIエンドポイント
- CSVファイル解析
- データベース操作

backendディレクトリから`go test ./...`でテストを実行。特定のテストを実行するには`-run`フラグを使用。