# Equity 計算バッチ処理

このディレクトリには、equity 計算を事前に行うバッチ処理の実装が含まれています。

## 概要

バッチ処理は以下の手順で動作します：

1. 6 つのシナリオ（プリセット）からランダムに 1 つ選択
2. 選択したシナリオに基づいてヒーローハンド、Opponent レンジ、フロップカードを生成
3. equity 計算を実行
4. 結果を DynamoDB に保存
5. 指定された回数または時間まで 1-4 を繰り返す

## 実行方法

バッチ処理を実行するには、以下のコマンドを使用します：

```bash
./scripts/run-batch.sh [オプション]
```

### オプション

- `-l <ファイル>` : ログファイル (デフォルト: stdout)
- `-d <ディレクトリ>` : データディレクトリ (デフォルト: data)
- `-D <YYYY-MM-DD>` : クイズの日付 (デフォルト: 明日)
- `-H <ホスト>` : PostgreSQL ホスト (デフォルト: localhost)
- `-p <ポート>` : PostgreSQL ポート (デフォルト: 5432)
- `-u <ユーザー>` : PostgreSQL ユーザー (デフォルト: postgres)
- `-P <パスワード>` : PostgreSQL パスワード (デフォルト: postgres)
- `-n <DB名>` : PostgreSQL データベース名 (デフォルト: plo_equity)
- `-h` : ヘルプを表示

### 使用例

```bash
# デフォルト設定で実行
./scripts/run-batch.sh

# 特定の日付のクイズを生成し、ログをファイルに出力
./scripts/run-batch.sh -D 2025-05-20 -l logs/batch.log

# カスタムPostgreSQLサーバーを指定
./scripts/run-batch.sh -H db.example.com -p 5433 -u myuser -P mypassword -n mydb
```

## 定期実行の設定

バッチ処理を定期的に実行するには、crontab を使用します。以下は設定例です：

### 毎日午前 3 時に実行

```bash
0 3 * * * cd /path/to/equity-distribution-app && ./scripts/run-batch.sh -n 500 -l logs/batch-$(date +\%Y\%m\%d).log
```

### 1 時間ごとに実行

```bash
0 * * * * cd /path/to/equity-distribution-app && ./scripts/run-batch.sh -n 100 -l logs/batch-$(date +\%Y\%m\%d-\%H).log
```

### crontab への追加方法

1. `crontab -e`コマンドで crontab を編集
2. 上記の設定例を追加
3. 保存して終了

## トラブルシューティング

### PostgreSQL に接続できない

エラーメッセージ：

```
エラー: PostgreSQLに接続できません。
```

解決策：

1. `docker-compose up -d`コマンドを実行して PostgreSQL を起動
2. PostgreSQL が起動していることを確認：`docker-compose ps`
3. PostgreSQL の接続情報が正しいか確認：`-H`, `-p`, `-u`, `-P`, `-n` オプション

### 権限エラー

エラーメッセージ：

```
permission denied: ./scripts/run-batch.sh
```

解決策：

```bash
chmod +x scripts/run-batch.sh
```

### メモリ不足エラー

エラーメッセージ：

```
fatal error: runtime: out of memory
```

解決策：

1. ワーカー数を減らす：`-w 2`オプションを使用
2. 計算回数を減らす：`-n 50`オプションを使用
