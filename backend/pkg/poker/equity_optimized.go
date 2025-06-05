package poker

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chehsunliu/poker"
)

// 計算精度の設定
const (
	FAST_ITERATIONS     = 1000  // 高速モード: ±3%誤差（95%信頼区間）
	NORMAL_ITERATIONS   = 5000  // 通常モード: ±1.4%誤差（95%信頼区間）
	ACCURATE_ITERATIONS = 10000 // 高精度モード: ±1%誤差（95%信頼区間）
)

// EquityCalculationConfig は計算設定を表します
type EquityCalculationConfig struct {
	MaxIterations    int
	TargetPrecision  float64
	MinIterations    int
	ConvergenceCheck int
}

// HandRankCache はハンドランクのキャッシュを管理します
type HandRankCache struct {
	cache sync.Map
	hits  int64
	total int64
}

var globalHandRankCache = &HandRankCache{}

// GetHandRank はキャッシュからハンドランクを取得します
func (hrc *HandRankCache) GetHandRank(hand []poker.Card, board []poker.Card) (int, bool) {
	key := generateHandKey(hand, board)
	hrc.total++
	if rank, ok := hrc.cache.Load(key); ok {
		hrc.hits++
		return rank.(int), true
	}
	return 0, false
}

// SetHandRank はハンドランクをキャッシュに保存します
func (hrc *HandRankCache) SetHandRank(hand []poker.Card, board []poker.Card, rank int) {
	key := generateHandKey(hand, board)
	hrc.cache.Store(key, rank)
}

// GetCacheStats はキャッシュの統計情報を返します
func (hrc *HandRankCache) GetCacheStats() (int64, int64, float64) {
	hits := hrc.hits
	total := hrc.total
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}
	return hits, total, hitRate
}

// generateHandKey はハンドとボードからキーを生成します
func generateHandKey(hand []poker.Card, board []poker.Card) string {
	var cards []string
	for _, card := range append(hand, board...) {
		cards = append(cards, card.String())
	}
	sort.Strings(cards) // 順序を正規化
	return strings.Join(cards, "")
}

// CalculateHandVsHandEquityMonteCarlo はモンテカルロシミュレーションでequityを計算します
func CalculateHandVsHandEquityMonteCarlo(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card, iterations int) (float64, error) {
	if HasCardDuplicates(yourHand, opponentHand, board) {
		return -1, fmt.Errorf("duplicate cards detected")
	}

	// 使用済みカードを除外
	usedCards := make(map[string]bool)
	for _, card := range append(append(yourHand, opponentHand...), board...) {
		usedCards[card.String()] = true
	}

	// 残りカードのデッキを作成
	deck := poker.NewDeck()
	fullDeck := deck.Draw(52)
	var remainingDeck []poker.Card
	for _, card := range fullDeck {
		if !usedCards[card.String()] {
			remainingDeck = append(remainingDeck, card)
		}
	}

	if len(remainingDeck) < 2 {
		return -1, fmt.Errorf("insufficient remaining cards")
	}

	wins := 0.0
	ties := 0.0

	// 乱数生成器を初期化
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// モンテカルロシミュレーション
	for i := 0; i < iterations; i++ {
		// ランダムに2枚選択（Fisher-Yatesシャッフルの簡略版）
		idx1 := rng.Intn(len(remainingDeck))
		idx2 := rng.Intn(len(remainingDeck) - 1)
		if idx2 >= idx1 {
			idx2++
		}

		finalBoard := append(board, remainingDeck[idx1], remainingDeck[idx2])

		winner := JudgeWinner(yourHand, opponentHand, finalBoard)
		switch winner {
		case "yourHand":
			wins++
		case "tie":
			ties += 0.5
		}
	}

	equity := (wins + ties) / float64(iterations) * 100
	return equity, nil
}

// CalculateHandVsHandEquityAdaptive は適応的精度制御でequityを計算します
func CalculateHandVsHandEquityAdaptive(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card, config EquityCalculationConfig) (float64, int, error) {
	if HasCardDuplicates(yourHand, opponentHand, board) {
		return -1, 0, fmt.Errorf("duplicate cards detected")
	}

	// 使用済みカードを除外
	usedCards := make(map[string]bool)
	for _, card := range append(append(yourHand, opponentHand...), board...) {
		usedCards[card.String()] = true
	}

	// 残りカードのデッキを作成
	deck := poker.NewDeck()
	fullDeck := deck.Draw(52)
	var remainingDeck []poker.Card
	for _, card := range fullDeck {
		if !usedCards[card.String()] {
			remainingDeck = append(remainingDeck, card)
		}
	}

	if len(remainingDeck) < 2 {
		return -1, 0, fmt.Errorf("insufficient remaining cards")
	}

	wins := 0.0
	total := 0.0
	var recentResults []float64

	// 乱数生成器を初期化
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < config.MaxIterations; i++ {
		// ランダムに2枚選択
		idx1 := rng.Intn(len(remainingDeck))
		idx2 := rng.Intn(len(remainingDeck) - 1)
		if idx2 >= idx1 {
			idx2++
		}

		finalBoard := append(board, remainingDeck[idx1], remainingDeck[idx2])

		winner := JudgeWinner(yourHand, opponentHand, finalBoard)
		switch winner {
		case "yourHand":
			wins++
		case "tie":
			wins += 0.5
		}
		total++

		// 収束チェック
		if i >= config.MinIterations && i%config.ConvergenceCheck == 0 {
			currentEquity := (wins / total) * 100
			recentResults = append(recentResults, currentEquity)

			if len(recentResults) >= 5 {
				// 最近の結果の標準偏差を計算
				if standardDeviation(recentResults) < config.TargetPrecision {
					return currentEquity, i + 1, nil
				}
				recentResults = recentResults[1:] // 古い結果を削除
			}
		}
	}

	return (wins / total) * 100, config.MaxIterations, nil
}

// CalculateHandVsRangeEquityMonteCarloParallel はモンテカルロシミュレーションで並列equity計算を行います
func CalculateHandVsRangeEquityMonteCarloParallel(yourHand []poker.Card, opponentHands [][]poker.Card, board []poker.Card) (map[string]float64, error) {
	equities := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 動的にイテレーション数を調整
	iterations := NORMAL_ITERATIONS
	if len(opponentHands) > 100 {
		iterations = FAST_ITERATIONS // 大量のハンドの場合は高速モード
		log.Printf("Using fast mode (%d iterations) for %d opponent hands", iterations, len(opponentHands))
	} else if len(opponentHands) < 20 {
		iterations = ACCURATE_ITERATIONS // 少数のハンドの場合は高精度モード
		log.Printf("Using accurate mode (%d iterations) for %d opponent hands", iterations, len(opponentHands))
	} else {
		log.Printf("Using normal mode (%d iterations) for %d opponent hands", iterations, len(opponentHands))
	}

	numCPU := runtime.NumCPU()
	semaphore := make(chan struct{}, numCPU)

	startTime := time.Now()

	for _, opponentHand := range opponentHands {
		if HasCardDuplicates(yourHand, opponentHand, board) {
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{}

		go func(currentOpponentHand []poker.Card) {
			defer wg.Done()
			defer func() { <-semaphore }()

			villainHandStr := ""
			for _, card := range currentOpponentHand {
				villainHandStr += card.String()
			}

			equity, err := CalculateHandVsHandEquityMonteCarlo(yourHand, currentOpponentHand, board, iterations)
			if err == nil && equity != -1 {
				mu.Lock()
				equities[villainHandStr] = equity
				mu.Unlock()
			}
		}(opponentHand)
	}

	wg.Wait()

	duration := time.Since(startTime)
	hits, total, hitRate := globalHandRankCache.GetCacheStats()
	log.Printf("Monte Carlo calculation completed in %v. Cache stats: %d/%d hits (%.1f%%)",
		duration, hits, total, hitRate)

	if len(equities) == 0 {
		return nil, fmt.Errorf("no valid equity calculations")
	}

	return equities, nil
}

// standardDeviation は標準偏差を計算します
func standardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// 平均を計算
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// 分散を計算
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values))

	return math.Sqrt(variance)
}

// GetDefaultAdaptiveConfig はデフォルトの適応的計算設定を返します
func GetDefaultAdaptiveConfig() EquityCalculationConfig {
	return EquityCalculationConfig{
		MaxIterations:    ACCURATE_ITERATIONS,
		TargetPrecision:  0.5, // 0.5%の精度
		MinIterations:    FAST_ITERATIONS,
		ConvergenceCheck: 200, // 200イテレーションごとに収束チェック
	}
}

// ClearHandRankCache はハンドランクキャッシュをクリアします
func ClearHandRankCache() {
	globalHandRankCache = &HandRankCache{}
}
