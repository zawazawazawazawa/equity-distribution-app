package poker

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/chehsunliu/poker"
)

// CalculateHandVsHandEquity calculates the equity between two hands
// Returns equity value and whether it was a cache hit
func CalculateHandVsHandEquity(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) (float64, bool) {
	// Check for duplicate cards
	if HasCardDuplicates(yourHand, opponentHand, board) {
		return -1, false // Return -1 to indicate invalid hand due to duplicate cards
	}

	// Generate the full deck
	deck := poker.NewDeck()
	fullDeck := deck.Draw(52) // Draw all 52 cards from the deck

	usedCards := append(yourHand, opponentHand...)
	usedCards = append(usedCards, board...)

	remainingDeck := []poker.Card{}
	for _, card := range fullDeck {
		used := false
		for _, usedCard := range usedCards {
			if card == usedCard {
				used = true
				break
			}
		}
		if !used {
			remainingDeck = append(remainingDeck, card)
		}
	}

	// Calculate equity since it wasn't found in DynamoDB
	totalOutcomes := 0.0
	winCount := 0.0

	for i := 0; i < len(remainingDeck); i++ {
		for j := i + 1; j < len(remainingDeck); j++ {
			finalBoard := append(board, remainingDeck[i], remainingDeck[j])
			winner := JudgeWinner(yourHand, opponentHand, finalBoard)
			if winner == "yourHand" {
				winCount += 1
			} else if winner == "tie" {
				winCount += 0.5
			}
			totalOutcomes += 1
		}
	}

	calculatedEquity := (winCount / totalOutcomes) * 100
	return calculatedEquity, false
}

// CalculateHandVsRangeEquityParallel は、1つのハンドと複数のハンドのレンジに対してエクイティを並列計算する
func CalculateHandVsRangeEquityParallel(yourHand []poker.Card, opponentHands [][]poker.Card, board []poker.Card) (map[string]float64, error) {
	// 結果を格納するマップ
	equities := make(map[string]float64)
	var mu sync.Mutex // 結果マップへのアクセスを保護するためのMutex
	var wg sync.WaitGroup

	numCPU := runtime.NumCPU()
	log.Printf("Using %d CPUs for parallel execution in CalculateHandVsRangeEquityParallel", numCPU)
	semaphore := make(chan struct{}, numCPU) // 同時実行数をCPUコア数に制限

	// 各オポーネントハンドに対してequity計算を並列で実行
	for _, opponentHand := range opponentHands {
		// カード重複チェック
		if HasCardDuplicates(yourHand, opponentHand, board) {
			continue
		}

		wg.Add(1)               // WaitGroupのカウンタをインクリメント
		semaphore <- struct{}{} // セマフォを取得（空きができるまでブロック）

		// goroutineでequity計算を実行
		go func(currentOpponentHand []poker.Card) {
			defer wg.Done()                // ゴルーチン完了時にカウンタをデクリメント
			defer func() { <-semaphore }() // セマフォを解放

			// ハンド文字列の生成
			villainHandStr := ""
			for _, card := range currentOpponentHand {
				villainHandStr += card.String()
			}

			// equity計算
			equity, _ := CalculateHandVsHandEquity(yourHand, currentOpponentHand, board)
			if equity != -1 {
				mu.Lock() // Mutexをロックしてequitiesマップを保護
				equities[villainHandStr] = equity
				mu.Unlock() // Mutexをアンロック
			}
		}(opponentHand)
	}

	wg.Wait() // すべてのゴルーチンが完了するのを待つ

	if len(equities) == 0 {
		return nil, fmt.Errorf("no valid equity calculations")
	}

	return equities, nil
}
