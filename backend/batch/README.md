# Equity 計算バッチ処理

このディレクトリには、equity 計算を事前に行うバッチ処理の実装が含まれています。

## 概要

バッチ処理は以下の手順で動作します：

1. 6 つのシナリオ（プリセット）からランダムに 1 つ選択
2. 選択したシナリオに基づいてヒーローハンド、オポーネントレンジ、フロップカードを生成
3. equity 計算を実行
4. 結果を DynamoDB に保存
5. 指定された回数または時間まで 1-4 を繰り返す

## 実行方法

バッチ処理を実行するには、以下のコマンドを使用します：

```bash
./scripts/run-batch.sh [オプション]
```

### オプション

- `-n <数値>` : 計算回数 (デフォルト: 100)
- `-w <数値>` : ワーカー数 (デフォルト: CPU コア数)
- `-e <URL>` : DynamoDB エンドポイント (デフォルト: http://localhost:4566)
- `-r <リージョン>` : AWS リージョン (デフォルト: us-east-1)
- `-l <ファイル>` : ログファイル (デフォルト: stdout)
- `-d <ディレクトリ>` : データディレクトリ (デフォルト: data)
- `-h` : ヘルプを表示

### 使用例

```bash
# デフォルト設定で実行
./scripts/run-batch.sh

# 1000回の計算を8ワーカーで実行し、ログをファイルに出力
./scripts/run-batch.sh -n 1000 -w 8 -l logs/batch.log

# カスタムDynamoDBエンドポイントを指定
./scripts/run-batch.sh -e http://dynamodb.example.com:8000
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

### DynamoDB に接続できない

エラーメッセージ：

```
エラー: DynamoDBに接続できません。docker-compose upを実行してDynamoDBを起動してください。
```

解決策：

1. `docker-compose up -d`コマンドを実行して DynamoDB を起動
2. DynamoDB が起動していることを確認：`docker-compose ps`

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
