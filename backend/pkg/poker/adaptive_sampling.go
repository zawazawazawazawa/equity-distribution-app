package poker

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/chehsunliu/poker"
)

// AdaptiveSamplingConfig は動的サンプリングの設定
type AdaptiveSamplingConfig struct {
	PilotSamples   int     // パイロットサンプル数
	MinSamples     int     // 最小サンプル数
	MaxSamples     int     // 最大サンプル数
	TargetError    float64 // 目標誤差（例: 0.01 = 1%）
	ConfidenceZ    float64 // 信頼区間のZ値（例: 1.96 = 95%）
	CheckInterval  int     // 収束チェック間隔
}

// DefaultAdaptiveConfig はデフォルトの動的サンプリング設定を返す
func DefaultAdaptiveConfig() AdaptiveSamplingConfig {
	return AdaptiveSamplingConfig{
		PilotSamples:  500,
		MinSamples:    1000,
		MaxSamples:    10000,
		TargetError:   0.01,
		ConfidenceZ:   1.96,
		CheckInterval: 100,
	}
}

// CalculateHandVsRangeAdaptive は動的サンプリングでエクイティを計算
func CalculateHandVsRangeAdaptive(
	yourHand []poker.Card,
	opponentRange [][]poker.Card,
	board []poker.Card,
	config AdaptiveSamplingConfig,
) (avgEquity float64, samplesUsed int, err error) {
	
	// 有効なレンジをフィルタリング
	var validRange [][]poker.Card
	for _, oppHand := range opponentRange {
		if !HasCardDuplicates(yourHand, oppHand, board) {
			validRange = append(validRange, oppHand)
		}
	}
	
	if len(validRange) == 0 {
		return 0, 0, fmt.Errorf("no valid opponent hands")
	}
	
	// 乱数生成器の初期化
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Phase 1: パイロットサンプリング
	pilotSize := config.PilotSamples
	if pilotSize > len(validRange) {
		pilotSize = len(validRange)
	}
	
	var equities []float64
	var sum, sumSquares float64
	
	// パイロットサンプルの実行
	for i := 0; i < pilotSize; i++ {
		oppHand := validRange[rng.Intn(len(validRange))]
		equity, _ := CalculateHandVsHandEquity(yourHand, oppHand, board)
		
		equities = append(equities, equity)
		sum += equity
		sumSquares += equity * equity
	}
	
	// Phase 2: 必要サンプル数の計算
	n := float64(len(equities))
	mean := sum / n
	variance := (sumSquares / n) - (mean * mean)
	stdDev := math.Sqrt(variance)
	
	// 必要サンプル数を計算（無限母集団）
	requiredN := math.Pow(config.ConfidenceZ*stdDev/config.TargetError, 2)
	
	// 有限母集団補正
	N := float64(len(validRange))
	requiredNFinite := requiredN / (1 + (requiredN-1)/N)
	
	// 必要サンプル数を範囲内にクリップ
	totalSamples := int(math.Ceil(requiredNFinite))
	if totalSamples < config.MinSamples {
		totalSamples = config.MinSamples
	}
	if totalSamples > config.MaxSamples {
		totalSamples = config.MaxSamples
	}
	if totalSamples > len(validRange) {
		totalSamples = len(validRange)
	}
	
	// Phase 3: 追加サンプリング（必要な場合）
	for i := len(equities); i < totalSamples; i++ {
		oppHand := validRange[rng.Intn(len(validRange))]
		equity, _ := CalculateHandVsHandEquity(yourHand, oppHand, board)
		
		equities = append(equities, equity)
		sum += equity
		sumSquares += equity * equity
		
		// 定期的に収束チェック
		if i%config.CheckInterval == 0 && i >= config.MinSamples {
			n := float64(i + 1)
			currentMean := sum / n
			currentVar := (sumSquares / n) - (currentMean * currentMean)
			stdError := math.Sqrt(currentVar / n)
			marginError := config.ConfidenceZ * stdError
			
			// 目標誤差を達成したら早期終了
			if marginError < config.TargetError {
				return currentMean, i + 1, nil
			}
		}
	}
	
	// 最終結果を返す
	finalMean := sum / float64(len(equities))
	return finalMean, len(equities), nil
}

// EstimateRequiredSamples はパイロットサンプルから必要サンプル数を推定
func EstimateRequiredSamples(
	pilotEquities []float64,
	rangeSize int,
	targetError float64,
	confidenceZ float64,
) int {
	if len(pilotEquities) == 0 {
		return 0
	}
	
	// 平均と分散を計算
	sum := 0.0
	for _, eq := range pilotEquities {
		sum += eq
	}
	mean := sum / float64(len(pilotEquities))
	
	variance := 0.0
	for _, eq := range pilotEquities {
		variance += math.Pow(eq-mean, 2)
	}
	variance /= float64(len(pilotEquities))
	stdDev := math.Sqrt(variance)
	
	// 必要サンプル数を計算
	n0 := math.Pow(confidenceZ*stdDev/targetError, 2)
	
	// 有限母集団補正
	N := float64(rangeSize)
	nFinite := n0 / (1 + (n0-1)/N)
	
	return int(math.Ceil(nFinite))
}