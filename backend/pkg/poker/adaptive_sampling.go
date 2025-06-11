package poker

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/chehsunliu/poker"
)

// AdaptiveSamplingConfig は適応的サンプリングの設定を表します
type AdaptiveSamplingConfig struct {
	MinSamples   int     // 最小サンプル数
	MaxSamples   int     // 最大サンプル数
	PilotSamples int     // パイロットサンプル数
	TargetError  float64 // 目標誤差
	ConfidenceZ  float64 // 信頼区間のZ値（95%信頼区間の場合1.96）
}

// DefaultAdaptiveConfig はデフォルトの適応的サンプリング設定を返します
func DefaultAdaptiveConfig() AdaptiveSamplingConfig {
	return AdaptiveSamplingConfig{
		MinSamples:   1000,
		MaxSamples:   10000,
		PilotSamples: 100,
		TargetError:  0.01,  // ±1%誤差
		ConfidenceZ:  1.96,  // 95%信頼区間
	}
}

// CalculateHandVsRangeAdaptiveWithDetails は動的サンプリングでエクイティを計算し、
// 各ハンドの個別エクイティも返す
func CalculateHandVsRangeAdaptiveWithDetails(
	yourHand []poker.Card,
	opponentRange [][]poker.Card,
	board []poker.Card,
	config AdaptiveSamplingConfig,
) (equities map[string]float64, avgEquity float64, samplesUsed int, err error) {
	
	// 結果を格納するマップ
	equities = make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	// 有効なレンジをフィルタリング
	var validRange [][]poker.Card
	handToString := make(map[string][]poker.Card) // ハンド文字列から実際のハンドへのマッピング
	
	for _, oppHand := range opponentRange {
		if !HasCardDuplicates(yourHand, oppHand, board) {
			validRange = append(validRange, oppHand)
			
			// ハンド文字列を生成
			handStr := ""
			for _, card := range oppHand {
				handStr += card.String()
			}
			handToString[handStr] = oppHand
		}
	}
	
	if len(validRange) == 0 {
		return nil, 0, 0, fmt.Errorf("no valid opponent hands")
	}
	
	log.Printf("Valid range size: %d hands", len(validRange))
	
	// Phase 1: パイロットサンプリングで必要サンプル数を推定
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	pilotSize := config.PilotSamples
	if pilotSize > len(validRange) {
		pilotSize = len(validRange)
	}
	
	var pilotEquities []float64
	var pilotSum, pilotSumSquares float64
	
	// パイロットサンプルの実行
	sampledIndices := make(map[int]bool)
	for i := 0; i < pilotSize; i++ {
		idx := rng.Intn(len(validRange))
		sampledIndices[idx] = true
		oppHand := validRange[idx]
		equity, _ := CalculateHandVsHandEquity(yourHand, oppHand, board)
		
		pilotEquities = append(pilotEquities, equity)
		pilotSum += equity
		pilotSumSquares += equity * equity
	}
	
	// 必要サンプル数の計算
	n := float64(len(pilotEquities))
	mean := pilotSum / n
	variance := (pilotSumSquares / n) - (mean * mean)
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
	
	log.Printf("Adaptive sampling: pilot mean=%.2f%%, stdDev=%.2f%%, required samples=%d",
		mean, stdDev, totalSamples)
	
	// Phase 2: サンプリングされたハンドに対して全数計算
	// 既にサンプリングされたインデックスを除いて追加サンプリング
	additionalSamples := totalSamples - len(sampledIndices)
	if additionalSamples > 0 {
		// 未サンプリングのインデックスリストを作成
		unsampled := []int{}
		for i := 0; i < len(validRange); i++ {
			if !sampledIndices[i] {
				unsampled = append(unsampled, i)
			}
		}
		
		// シャッフルして必要数だけ選択
		for i := range unsampled {
			j := rng.Intn(i + 1)
			unsampled[i], unsampled[j] = unsampled[j], unsampled[i]
		}
		
		for i := 0; i < additionalSamples && i < len(unsampled); i++ {
			sampledIndices[unsampled[i]] = true
		}
	}
	
	// Phase 3: サンプリングされた全ハンドに対して並列でエクイティ計算
	numCPU := runtime.NumCPU()
	semaphore := make(chan struct{}, numCPU)
	
	totalEquity := 0.0
	samplesUsed = 0
	
	for idx := range sampledIndices {
		wg.Add(1)
		semaphore <- struct{}{}
		
		go func(handIdx int) {
			defer wg.Done()
			defer func() { <-semaphore }()
			
			oppHand := validRange[handIdx]
			
			// ハンド文字列を生成
			handStr := ""
			for _, card := range oppHand {
				handStr += card.String()
			}
			
			// 全ターン・リバーでエクイティ計算（全数計算）
			equity, _ := CalculateHandVsHandEquity(yourHand, oppHand, board)
			
			if equity != -1 {
				mu.Lock()
				equities[handStr] = equity
				totalEquity += equity
				samplesUsed++
				mu.Unlock()
			}
		}(idx)
	}
	
	wg.Wait()
	
	// 平均エクイティを計算
	if samplesUsed > 0 {
		avgEquity = totalEquity / float64(samplesUsed)
	}
	
	log.Printf("Adaptive sampling completed: sampled %d hands out of %d total hands (%.1f%%)",
		samplesUsed, len(validRange), float64(samplesUsed)/float64(len(validRange))*100)
	
	return equities, avgEquity, samplesUsed, nil
}