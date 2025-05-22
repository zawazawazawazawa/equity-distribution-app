# 画像アップロードテストスクリプト

このスクリプトは、`equity-distribution-backend`の画像生成と R2 ストレージへのアップロード機能をテストするためのものです。バッチ処理の他の部分（エクイティ計算やデータベース処理など）を含まず、画像生成とアップロードのみをテストします。

## 機能

- 指定されたパラメータ（日付、シナリオ名、ヒーローハンド、フロップ）に基づいて画像を生成
- 生成した画像を R2 ストレージにアップロード
- 画像の公開 URL を生成

## 必要な環境変数

R2 ストレージへのアップロードをテストするには、以下の環境変数を設定する必要があります：

```bash
export R2_ENDPOINT="your-r2-endpoint"
export R2_ACCESS_KEY="your-r2-access-key"
export R2_SECRET_KEY="your-r2-secret-key"
export R2_BUCKET="your-r2-bucket"
```

これらの環境変数が設定されていない場合、スクリプトは画像生成のみを行い、アップロードはスキップされます。

## 実行方法

### 直接実行

```bash
# バックエンドディレクトリに移動
cd backend

# ビルド
go build -o test-image-upload cmd/test-image-upload/main.go

# 実行（デフォルトパラメータ）
./test-image-upload

# 実行（パラメータ指定）
./test-image-upload -date=2023-05-22 -scenario="SRP UTG vs BB" -hand=AsKsQsJs -flop=2c3d4h
```

### ヘルパースクリプトを使用

より簡単に実行するために、ヘルパースクリプト `scripts/test-image-upload.sh` を使用できます：

```bash
# 実行（デフォルトパラメータ）
./scripts/test-image-upload.sh

# 実行（パラメータ指定）
./scripts/test-image-upload.sh --date 2023-05-22 --scenario "SRP UTG vs BB" --hand AsKsQsJs --flop 2c3d4h

# ヘルプを表示
./scripts/test-image-upload.sh --help
```

## パラメータ

| パラメータ                 | 説明                    | デフォルト値    |
| -------------------------- | ----------------------- | --------------- |
| `-date` / `--date`         | 日付（YYYY-MM-DD 形式） | 今日の日付      |
| `-scenario` / `--scenario` | シナリオ名              | "SRP UTG vs BB" |
| `-hand` / `--hand`         | ヒーローハンド          | "AsKsQsJs"      |
| `-flop` / `--flop`         | フロップカード          | "2c3d4h"        |
| `-log` / `--log`           | ログファイル            | （標準出力）    |

## 出力

スクリプトは以下の情報を出力します：

1. 使用するパラメータ（日付、シナリオ名、ヒーローハンド、フロップ）
2. 画像生成の結果
3. R2 へのアップロード結果（環境変数が設定されている場合）
4. 画像の公開 URL（環境変数が設定されている場合）

## トラブルシューティング

### 画像生成に失敗する場合

- フォントファイルが `backend/fonts/` ディレクトリに存在することを確認してください
- カード画像が `backend/images/playing_cards/` ディレクトリに存在することを確認してください

### R2 へのアップロードに失敗する場合

- 環境変数が正しく設定されていることを確認してください
- R2 のアクセス権限を確認してください
- ネットワーク接続を確認してください

## 注意事項

- このスクリプトは、`equity-distribution-backend`のパッケージに依存しています
- 画像は `backend/images/daily-quiz/` ディレクトリに生成されます
