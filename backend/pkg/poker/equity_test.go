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

func TestCalculateHandVsRangeEquityParallel(t *testing.T) {
	// テストケース1: 通常のエクイティ計算
	t.Run("Normal equity calculation", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
			poker.NewCard("Kc"),
			poker.NewCard("Qc"),
		}
		opponentHands := [][]poker.Card{
			{
				poker.NewCard("Kh"),
				poker.NewCard("Kd"),
				poker.NewCard("Jc"),
				poker.NewCard("Tc"),
			},
			{
				poker.NewCard("Qh"),
				poker.NewCard("Qd"),
				poker.NewCard("Jd"),
				poker.NewCard("Td"),
			},
		}
		board := []poker.Card{
			poker.NewCard("2c"),
			poker.NewCard("7d"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		equities, err := CalculateHandVsRangeEquityParallel(yourHand, opponentHands, board)

		// 結果の検証
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(equities) != 2 {
			t.Errorf("Expected 2 equity results, got %d", len(equities))
		}
		for villainHand, equity := range equities {
			if equity < 0 {
				t.Errorf("Expected positive equity for %s, got %.2f", villainHand, equity)
			}
		}
	})

	// テストケース2: カードの重複がある場合
	t.Run("Duplicate cards", func(t *testing.T) {
		// テストデータの準備 - 重複するカードを含む
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
			poker.NewCard("Kc"),
			poker.NewCard("Qc"),
		}
		opponentHands := [][]poker.Card{
			{
				poker.NewCard("Ah"), // 重複するカード
				poker.NewCard("Kd"),
				poker.NewCard("Jc"),
				poker.NewCard("Tc"),
			},
			{
				poker.NewCard("Qh"),
				poker.NewCard("Qd"),
				poker.NewCard("Jd"),
				poker.NewCard("Td"),
			},
		}
		board := []poker.Card{
			poker.NewCard("2c"),
			poker.NewCard("7d"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		equities, err := CalculateHandVsRangeEquityParallel(yourHand, opponentHands, board)

		// 結果の検証
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		// 重複するカードを含むハンドは除外されるため、結果は1つだけになるはず
		if len(equities) != 1 {
			t.Errorf("Expected 1 equity result after filtering duplicates, got %d", len(equities))
		}
	})

	// テストケース3: 空のレンジの場合
	t.Run("Empty range", func(t *testing.T) {
		// テストデータの準備 - 空のレンジ
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
			poker.NewCard("Kc"),
			poker.NewCard("Qc"),
		}
		opponentHands := [][]poker.Card{}
		board := []poker.Card{
			poker.NewCard("2c"),
			poker.NewCard("7d"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		_, err := CalculateHandVsRangeEquityParallel(yourHand, opponentHands, board)

		// 結果の検証
		if err == nil {
			t.Error("Expected error for empty range, got nil")
		}
	})

	// テストケース4: すべてのハンドが重複する場合
	t.Run("All hands have duplicates", func(t *testing.T) {
		// テストデータの準備 - すべてのハンドが重複
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
			poker.NewCard("Kc"),
			poker.NewCard("Qc"),
		}
		opponentHands := [][]poker.Card{
			{
				poker.NewCard("Ah"), // 重複するカード
				poker.NewCard("Kd"),
				poker.NewCard("Jc"),
				poker.NewCard("Tc"),
			},
			{
				poker.NewCard("Ad"), // 重複するカード
				poker.NewCard("Qd"),
				poker.NewCard("Jd"),
				poker.NewCard("Td"),
			},
		}
		board := []poker.Card{
			poker.NewCard("2c"),
			poker.NewCard("7d"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		_, err := CalculateHandVsRangeEquityParallel(yourHand, opponentHands, board)

		// 結果の検証
		if err == nil {
			t.Error("Expected error when all hands have duplicates, got nil")
		}
	})

	// テストケース5: 大量のハンドに対する並列処理
	t.Run("Large number of hands", func(t *testing.T) {
		// テストデータの準備 - 多数のハンド
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
			poker.NewCard("Kc"),
			poker.NewCard("Qc"),
		}

		// 10個のハンドを生成
		opponentHands := [][]poker.Card{}
		suits := []string{"h", "d", "c", "s"}
		ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}

		// 重複しないようにハンドを生成
		for i := 0; i < 10; i++ {
			hand := []poker.Card{
				poker.NewCard(ranks[(i*4)%13] + suits[i%4]),
				poker.NewCard(ranks[(i*4+1)%13] + suits[(i+1)%4]),
				poker.NewCard(ranks[(i*4+2)%13] + suits[(i+2)%4]),
				poker.NewCard(ranks[(i*4+3)%13] + suits[(i+3)%4]),
			}
			opponentHands = append(opponentHands, hand)
		}

		board := []poker.Card{
			poker.NewCard("2c"),
			poker.NewCard("7d"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		equities, err := CalculateHandVsRangeEquityParallel(yourHand, opponentHands, board)

		// 結果の検証
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(equities) != 10 {
			t.Errorf("Expected 10 equity results, got %d", len(equities))
		}
	})
}
