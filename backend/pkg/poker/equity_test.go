package poker

import (
	"testing"

	"github.com/chehsunliu/poker"
)

func TestCalculateHandVsHandEquity(t *testing.T) {
	// テストケース1: 通常のエクイティ計算
	t.Run("Normal equity calculation", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Kd"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		equity, cacheHit := CalculateHandVsHandEquity(yourHand, opponentHand, board)

		// 結果の検証
		if equity < 0 {
			t.Errorf("Expected positive equity, got %.2f", equity)
		}
		if cacheHit {
			t.Error("Expected cache miss, got cache hit")
		}
	})

	// テストケース2: カードの重複がある場合
	t.Run("Duplicate cards", func(t *testing.T) {
		// テストデータの準備 - 重複するカードを含む
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Ah"), // 重複するカード
			poker.NewCard("Kd"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		equity, cacheHit := CalculateHandVsHandEquity(yourHand, opponentHand, board)

		// 結果の検証
		if equity != -1 {
			t.Errorf("Expected -1 for duplicate cards, got %.2f", equity)
		}
		if cacheHit {
			t.Error("Expected cache miss, got cache hit")
		}
	})

	// テストケース3: ボードとハンドの間でカードが重複する場合
	t.Run("Duplicate cards between hand and board", func(t *testing.T) {
		// テストデータの準備 - ボードとハンドの間で重複するカード
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Kd"),
		}
		board := []poker.Card{
			poker.NewCard("Ah"), // 重複するカード
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		equity, cacheHit := CalculateHandVsHandEquity(yourHand, opponentHand, board)

		// 結果の検証
		if equity != -1 {
			t.Errorf("Expected -1 for duplicate cards, got %.2f", equity)
		}
		if cacheHit {
			t.Error("Expected cache miss, got cache hit")
		}
	})

	// テストケース4: 特定のハンドの組み合わせでのエクイティ検証
	t.Run("Specific hand combination equity", func(t *testing.T) {
		// テストデータの準備 - AA vs KK on dry board
		yourHand := []poker.Card{
			poker.NewCard("As"),
			poker.NewCard("Ac"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Ks"),
			poker.NewCard("Kc"),
		}
		board := []poker.Card{
			poker.NewCard("2h"),
			poker.NewCard("7d"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		equity, _ := CalculateHandVsHandEquity(yourHand, opponentHand, board)

		// 結果の検証 - AA vs KK on dry board should have equity around 80-95%
		if equity < 75 || equity > 95 {
			t.Errorf("Expected equity for AA vs KK on dry board to be around 80-95%%, got %.2f%%", equity)
		}
	})

	// テストケース5: 最小限のボードでのエクイティ計算（3枚のカード）
	t.Run("Minimal board", func(t *testing.T) {
		// テストデータの準備 - 3枚のボード（最小限必要な枚数）
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Kd"),
		}
		board := []poker.Card{
			poker.NewCard("2h"),
			poker.NewCard("7d"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		equity, cacheHit := CalculateHandVsHandEquity(yourHand, opponentHand, board)

		// 結果の検証
		if equity < 0 {
			t.Errorf("Expected positive equity, got %.2f", equity)
		}
		if cacheHit {
			t.Error("Expected cache miss, got cache hit")
		}
	})
}
