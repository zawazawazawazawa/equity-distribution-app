package main

import (
	"fmt"
	"log"
	"time"

	"github.com/chehsunliu/poker"

	"equity-distribution-backend/pkg/db"
	"equity-distribution-backend/pkg/image"
)

func main() {
	// 1. PostgreSQL接続の設定
	pgConfig := db.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "plo_equity",
	}

	// 2. PostgreSQLに接続
	pgDB, err := db.GetPostgresConnection(pgConfig)
	if err != nil {
		log.Fatalf("PostgreSQLへの接続に失敗しました: %v", err)
	}
	defer pgDB.Close()

	// 3. 2025-05-15の日付を指定
	targetDate, err := time.Parse("2006-01-02", "2025-05-15")
	if err != nil {
		log.Fatalf("日付の解析に失敗しました: %v", err)
	}

	// 4. 指定された日付のデータを取得
	results, err := db.GetDailyQuizResultsByDate(pgDB, targetDate)
	if err != nil {
		log.Fatalf("データの取得に失敗しました: %v", err)
	}

	if len(results) == 0 {
		log.Fatalf("2025-05-15のデータが見つかりませんでした")
	}

	// 5. 最初のレコードを取得
	firstResult := results[0]

	// 6. データを取得
	scenario := firstResult["scenario"].(string)
	heroHand := firstResult["hero_hand"].(string)
	flopStr := firstResult["flop"].(string)

	// 7. フロップの文字列をpoker.Card型の配列に変換
	var flop []poker.Card
	for i := 0; i < len(flopStr); i += 2 {
		if i+1 >= len(flopStr) {
			break
		}
		rank := string(flopStr[i])
		suit := string(flopStr[i+1])
		cardStr := rank + suit
		card := poker.NewCard(cardStr)
		flop = append(flop, card)
	}

	// 8. 画像生成
	err = image.GenerateDailyQuizImage(targetDate, scenario, heroHand, flop)
	if err != nil {
		log.Fatalf("画像生成に失敗しました: %v", err)
	}

	// 9. 成功メッセージを表示
	fmt.Printf("画像が正常に生成されました: ./images/daily-quiz/%s.png\n", targetDate.Format("2006-01-02"))
}
