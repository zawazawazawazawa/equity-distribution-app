package poker

import (
	"testing"

	"github.com/chehsunliu/poker"
)

func TestJudgeWinner(t *testing.T) {
	// テストケース1: Hold'em - yourHandが勝つ場合
	t.Run("Hold'em - YourHand wins", func(t *testing.T) {
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
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "yourHand" {
			t.Errorf("Expected 'yourHand' to win, got %s", result)
		}
	})

	// テストケース2: Hold'em - opponentHandが勝つ場合
	t.Run("Hold'em - OpponentHand wins", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Kd"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "opponentHand" {
			t.Errorf("Expected 'opponentHand' to win, got %s", result)
		}
	})

	// テストケース3: Hold'em - 引き分けの場合
	t.Run("Hold'em - Tie", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kd"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Ad"),
			poker.NewCard("Kh"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "tie" {
			t.Errorf("Expected 'tie', got %s", result)
		}
	})

	// テストケース4: Hold'em - フラッシュ vs ストレート
	t.Run("Hold'em - Flush vs Straight", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kh"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("9d"),
			poker.NewCard("8c"),
		}
		board := []poker.Card{
			poker.NewCard("Qh"),
			poker.NewCard("Jh"),
			poker.NewCard("Th"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "yourHand" {
			t.Errorf("Expected 'yourHand' to win with flush, got %s", result)
		}
	})

	// テストケース5: Hold'em - フルハウス vs フラッシュ
	t.Run("Hold'em - Full House vs Flush", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Qh"),
		}
		board := []poker.Card{
			poker.NewCard("Ac"),
			poker.NewCard("Kd"),
			poker.NewCard("Kc"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "yourHand" {
			t.Errorf("Expected 'yourHand' to win with full house, got %s", result)
		}
	})
}

func TestJudgeWinnerPLO(t *testing.T) {
	// テストケース1: PLO - 正常なフラッシュ vs フラッシュ
	t.Run("PLO - Valid flush vs flush", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kh"), // ハートが2枚
			poker.NewCard("2c"),
			poker.NewCard("3d"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Qh"),
			poker.NewCard("Jh"), // ハートが2枚
			poker.NewCard("4c"),
			poker.NewCard("5d"),
		}
		board := []poker.Card{
			poker.NewCard("Th"),
			poker.NewCard("9h"),
			poker.NewCard("8h"),
			poker.NewCard("7c"),
			poker.NewCard("6s"),
		}

		// 関数の実行
		result := JudgeWinnerPLO(yourHand, opponentHand, board)

		// 結果の検証 - 実際の最高の組み合わせを確認
		// yourHand: A♥K♥ + T♥9♥8♥ = A♥K♥T♥9♥8♥ (Ace high flush)
		// opponentHand: Q♥J♥ + T♥9♥8♥ = Q♥J♥T♥9♥8♥ (Straight flush 8-Q!)
		// opponentHandがストレートフラッシュで勝つべき
		if result != "opponentHand" {
			t.Errorf("Expected 'opponentHand' to win with straight flush (Q♥J♥T♥9♥8♥ vs A♥K♥T♥9♥8♥), got %s", result)
		}
	})

	// テストケース2: PLO - 1枚フラッシュの防止テスト
	t.Run("PLO - Prevent one-card flush", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"), // ハートは1枚のみ
			poker.NewCard("2c"),
			poker.NewCard("3d"),
			poker.NewCard("4s"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Qh"), // ハートが2枚
			poker.NewCard("5c"),
			poker.NewCard("6d"),
		}
		board := []poker.Card{
			poker.NewCard("Jh"),
			poker.NewCard("Th"),
			poker.NewCard("9h"),
			poker.NewCard("8h"),
			poker.NewCard("7c"),
		}

		// 関数の実行
		result := JudgeWinnerPLO(yourHand, opponentHand, board)

		// 結果の検証 - opponentHandがフラッシュで勝つべき
		// yourHandは手札にハートが1枚しかないためフラッシュにならない
		if result != "opponentHand" {
			t.Errorf("Expected 'opponentHand' to win with flush (yourHand cannot make flush with only 1 heart), got %s", result)
		}
	})

	// テストケース3: PLO - 0枚フラッシュの防止テスト
	t.Run("PLO - Prevent zero-card flush", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("2c"), // ハートなし
			poker.NewCard("3d"),
			poker.NewCard("4s"),
			poker.NewCard("5c"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Qh"), // ハートが2枚
			poker.NewCard("6c"),
			poker.NewCard("7d"),
		}
		board := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Jh"),
			poker.NewCard("Th"),
			poker.NewCard("9h"),
			poker.NewCard("8h"),
		}

		// 関数の実行
		result := JudgeWinnerPLO(yourHand, opponentHand, board)

		// 結果の検証 - opponentHandがフラッシュで勝つべき
		// yourHandは手札にハートがないためフラッシュにならない
		if result != "opponentHand" {
			t.Errorf("Expected 'opponentHand' to win with flush (yourHand has no hearts), got %s", result)
		}
	})

	// テストケース4: PLO - ストレートの正しい判定
	t.Run("PLO - Valid straight", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("9h"),
			poker.NewCard("8c"), // 9,8でストレート可能
			poker.NewCard("2d"),
			poker.NewCard("3s"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"), // ペア
			poker.NewCard("4c"),
			poker.NewCard("5d"),
		}
		board := []poker.Card{
			poker.NewCard("Th"),
			poker.NewCard("Jd"),
			poker.NewCard("Qc"),
			poker.NewCard("Ks"),
			poker.NewCard("2h"),
		}

		// 関数の実行
		result := JudgeWinnerPLO(yourHand, opponentHand, board)

		// 結果の検証 - yourHandがストレート(9-K)で勝つべき
		if result != "yourHand" {
			t.Errorf("Expected 'yourHand' to win with straight, got %s", result)
		}
	})

	// テストケース5: PLO - 手札2枚必須の確認
	t.Run("PLO - Must use exactly 2 hole cards", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("2h"),
			poker.NewCard("3c"), // 低いカード
			poker.NewCard("4d"),
			poker.NewCard("5s"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("6h"),
			poker.NewCard("7c"), // 少し高いカード
			poker.NewCard("8d"),
			poker.NewCard("9s"),
		}
		board := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kd"),
			poker.NewCard("Qc"),
			poker.NewCard("Js"),
			poker.NewCard("Th"), // ボードだけでロイヤルストレートフラッシュ
		}

		// 関数の実行
		result := JudgeWinnerPLO(yourHand, opponentHand, board)

		// 結果の検証 - opponentHandが勝つべき（手札2枚を使う必要があるため、ボードの役は使えない）
		if result != "opponentHand" {
			t.Errorf("Expected 'opponentHand' to win (board royal flush cannot be used without 2 hole cards), got %s", result)
		}
	})
}

func TestJudgeWinnerPLO5(t *testing.T) {
	// テストケース1: PLO5 - 正常なフラッシュ vs フラッシュ
	t.Run("PLO5 - Valid flush vs flush", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kh"), // ハートが2枚
			poker.NewCard("2c"),
			poker.NewCard("3d"),
			poker.NewCard("4s"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Qh"),
			poker.NewCard("Jh"), // ハートが2枚
			poker.NewCard("5c"),
			poker.NewCard("6d"),
			poker.NewCard("7s"),
		}
		board := []poker.Card{
			poker.NewCard("Th"),
			poker.NewCard("9h"),
			poker.NewCard("8h"),
			poker.NewCard("7c"),
			poker.NewCard("6c"),
		}

		// 関数の実行
		result := JudgeWinnerPLO5(yourHand, opponentHand, board)

		// 結果の検証 - opponentHandがストレートフラッシュで勝つべき
		if result != "opponentHand" {
			t.Errorf("Expected 'opponentHand' to win with straight flush, got %s", result)
		}
	})

	// テストケース2: PLO5 - Auto-detect 5 card hands
	t.Run("PLO5 - Auto detect 5 card hands", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
			poker.NewCard("Kh"),
			poker.NewCard("Kd"),
			poker.NewCard("Qs"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("2h"),
			poker.NewCard("3h"),
			poker.NewCard("4c"),
			poker.NewCard("5d"),
			poker.NewCard("6s"),
		}
		board := []poker.Card{
			poker.NewCard("Ac"),
			poker.NewCard("Ks"),
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Th"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証 - yourHandがフルハウスで勝つべき
		if result != "yourHand" {
			t.Errorf("Expected 'yourHand' to win with full house, got %s", result)
		}
	})

	// テストケース3: PLO5 - より多くの組み合わせからベストハンドを選択
	t.Run("PLO5 - More combinations test", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("9h"),
			poker.NewCard("8h"),
			poker.NewCard("7h"),
			poker.NewCard("6h"),
			poker.NewCard("5h"), // ストレートフラッシュ可能
		}
		opponentHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
			poker.NewCard("As"),
			poker.NewCard("Kc"),
			poker.NewCard("Kd"), // スリーカード可能
		}
		board := []poker.Card{
			poker.NewCard("Ac"),
			poker.NewCard("Kh"),
			poker.NewCard("Qd"),
			poker.NewCard("Jh"),
			poker.NewCard("Th"),
		}

		// 関数の実行
		result := JudgeWinnerPLO5(yourHand, opponentHand, board)

		// 結果の検証 - yourHandがストレートフラッシュで勝つべき
		if result != "yourHand" {
			t.Errorf("Expected 'yourHand' to win with straight flush, got %s", result)
		}
	})

	// テストケース4: PLO5 - 1枚フラッシュの防止テスト
	t.Run("PLO5 - Prevent one-card flush", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"), // ハートは1枚のみ
			poker.NewCard("2c"),
			poker.NewCard("3d"),
			poker.NewCard("4s"),
			poker.NewCard("5c"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Qh"), // ハートが2枚
			poker.NewCard("6c"),
			poker.NewCard("7d"),
			poker.NewCard("8s"),
		}
		board := []poker.Card{
			poker.NewCard("Jh"),
			poker.NewCard("Th"),
			poker.NewCard("9h"),
			poker.NewCard("8h"),
			poker.NewCard("7c"),
		}

		// 関数の実行
		result := JudgeWinnerPLO5(yourHand, opponentHand, board)

		// 結果の検証 - opponentHandがフラッシュで勝つべき
		if result != "opponentHand" {
			t.Errorf("Expected 'opponentHand' to win with flush, got %s", result)
		}
	})

	// テストケース5: PLO5 - タイの判定
	t.Run("PLO5 - Tie game", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kd"),
			poker.NewCard("2c"),
			poker.NewCard("3s"),
			poker.NewCard("4h"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Ad"),
			poker.NewCard("Kh"),
			poker.NewCard("5c"),
			poker.NewCard("6s"),
			poker.NewCard("7h"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
			poker.NewCard("9h"),
			poker.NewCard("8c"),
		}

		// 関数の実行
		result := JudgeWinnerPLO5(yourHand, opponentHand, board)

		// 結果の検証 - 両方ともA-Kでストレートなのでタイ
		if result != "tie" {
			t.Errorf("Expected 'tie', got %s", result)
		}
	})
}