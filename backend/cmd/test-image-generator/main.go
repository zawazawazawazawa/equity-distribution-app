package main

import (
	"fmt"
	"log"
	"time"

	"github.com/chehsunliu/poker"
	"equity-distribution-backend/pkg/image"
)

func main() {
	fmt.Println("画像生成テストを開始します...")

	// テスト用の日付（今日の日付）
	date := time.Now()

	// 4-card PLOのテスト
	fmt.Println("\n4-card PLO画像を生成中...")
	scenario4card := "UTG Open, BTN 3-bet, UTG Call"
	heroHand4card := "AsKsQcJc"
	flop4card := []poker.Card{
		poker.NewCard("Ts"),
		poker.NewCard("9s"),
		poker.NewCard("8h"),
	}

	err := image.GenerateDailyQuizImage(date, scenario4card, heroHand4card, flop4card)
	if err != nil {
		log.Fatalf("4-card PLO画像生成エラー: %v", err)
	}
	fmt.Println("4-card PLO画像が正常に生成されました")

	// 5-card PLOのテスト
	fmt.Println("\n5-card PLO画像を生成中...")
	scenario5card := "BTN Open, BB 3-bet, BTN Call"
	heroHand5card := "AsKsQsJsTs"
	flop5card := []poker.Card{
		poker.NewCard("9s"),
		poker.NewCard("8c"),
		poker.NewCard("7d"),
	}

	err = image.GenerateDailyQuizImage(date, scenario5card, heroHand5card, flop5card)
	if err != nil {
		log.Fatalf("5-card PLO画像生成エラー: %v", err)
	}
	fmt.Println("5-card PLO画像が正常に生成されました")

	fmt.Println("\nすべての画像が正常に生成されました！")
	fmt.Printf("生成された画像は以下のパスにあります:\n")
	fmt.Printf("- 4-card PLO: ./images/daily-quiz/4card/%s.png\n", date.Format("2006-01-02"))
	fmt.Printf("- 5-card PLO: ./images/daily-quiz/5card/%s.png\n", date.Format("2006-01-02"))
}