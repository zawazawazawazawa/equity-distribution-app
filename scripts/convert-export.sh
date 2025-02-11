#!/bin/bash

# スクリプトのディレクトリに移動
cd "$(dirname "$0")"

# export.jsonからimport-items.jsonを生成
jq '{
  PloEquity: [
    .Items[] | {
      PutRequest: {
        Item: .
      }
    }
  ]
}' ../export.json > ../import-items.json

echo "Conversion completed: import-items.json has been created"
