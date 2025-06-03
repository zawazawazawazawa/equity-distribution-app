# Database Migrations

このディレクトリには、データベーススキーマの変更を管理するためのマイグレーションファイルが含まれています。

## ファイル命名規則

マイグレーションファイルは以下の命名規則に従います：

```
{version}_{description}.{direction}.sql
```

- `version`: 6 桁の数字（例: 000001）
- `description`: マイグレーションの説明（スネークケース）
- `direction`: `up`（適用）または `down`（ロールバック）

例：

- `000001_create_daily_quiz_results.up.sql`
- `000001_create_daily_quiz_results.down.sql`

## 使用方法

### 1. CLI コマンドを直接実行

```bash
# backendディレクトリに移動
cd backend

# すべてのマイグレーションを適用
go run cmd/migrate/main.go up

# 最後のマイグレーションをロールバック
go run cmd/migrate/main.go down

# 現在のマイグレーション状態を確認
go run cmd/migrate/main.go status

# 特定のバージョンまで適用
go run cmd/migrate/main.go version 1

# データベースを特定のバージョンに強制設定（注意して使用）
go run cmd/migrate/main.go force 1
```

### 2. 便利スクリプトを使用

```bash
# プロジェクトルートから実行

# マイグレーション適用
./scripts/migrate-up.sh

# ロールバック
./scripts/migrate-down.sh

# ステータス確認
./scripts/migrate-status.sh
```

## 環境変数

以下の環境変数でデータベース接続を設定できます：

- `POSTGRES_HOST` (デフォルト: localhost)
- `POSTGRES_PORT` (デフォルト: 5432)
- `POSTGRES_USER` (デフォルト: postgres)
- `POSTGRES_PASSWORD` (デフォルト: postgres)
- `POSTGRES_DBNAME` (デフォルト: plo_equity)

## 新しいマイグレーションの作成

1. 新しいマイグレーションファイルを作成：

   ```bash
   # 例: テーブルにカラムを追加する場合
   touch backend/migrations/000002_add_column_to_table.up.sql
   touch backend/migrations/000002_add_column_to_table.down.sql
   ```

2. `.up.sql`ファイルに変更を記述：

   ```sql
   ALTER TABLE daily_quiz_results ADD COLUMN new_column VARCHAR(255);
   ```

3. `.down.sql`ファイルにロールバック処理を記述：
   ```sql
   ALTER TABLE daily_quiz_results DROP COLUMN new_column;
   ```

## 注意事項

- **本番環境での実行前には必ずバックアップを取得してください**
- マイグレーションは順番に実行されるため、バージョン番号を正しく設定してください
- `down.sql`ファイルは必ず作成し、`up.sql`の変更を正確に元に戻すように記述してください
- 本番環境では`force`コマンドの使用は避けてください

## トラブルシューティング

### データベースが"dirty"状態になった場合

```bash
# 現在の状態を確認
go run cmd/migrate/main.go status

# 適切なバージョンに強制設定（注意深く実行）
go run cmd/migrate/main.go force <version>
```

### マイグレーションが失敗した場合

1. エラーメッセージを確認
2. データベースの状態を確認
3. 必要に応じて手動でデータベースを修正
4. `force`コマンドで状態をリセット
