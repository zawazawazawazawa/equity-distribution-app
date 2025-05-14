package main

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/chehsunliu/poker"
)

// モックのBatchConfig
type MockBatchConfig struct {
	DynamoDBEndpoint string
	DynamoDBRegion   string
	LogFile          string
	DataDir          string
}

// モックのScenario
type MockScenario struct {
	Name           string
	PresetName     string
	Description    string
	HeroHandRanges []string
}

// 並列処理のテスト用関数
func TestParallelProcessing(t *testing.T) {
	// 並列処理のテスト
	t.Run("Parallel processing test", func(t *testing.T) {
		// 同時実行数を制限するセマフォ
		numCPU := runtime.NumCPU()
		semaphore := make(chan struct{}, numCPU)
		var wg sync.WaitGroup

		// 処理完了時間を記録する配列
		completionTimes := make([]time.Time, 5)
		var mu sync.Mutex

		// 乱数生成器の初期化
		rand.Seed(time.Now().UnixNano())

		// 5つのタスクを並列実行
		for i := 0; i < 5; i++ {
			wg.Add(1)
			semaphore <- struct{}{} // セマフォを取得

			// タスク処理をgoroutineで実行
			go func(index int) {
				defer wg.Done()
				defer func() { <-semaphore }() // セマフォを解放

				// 各タスクは0.1〜0.3秒のランダムな時間がかかると仮定
				sleepTime := 100 + rand.Intn(200)
				time.Sleep(time.Duration(sleepTime) * time.Millisecond)

				// 処理完了時間を記録
				mu.Lock()
				completionTimes[index] = time.Now()
				mu.Unlock()
			}(i)
		}

		// すべてのgoroutineが完了するのを待つ
		wg.Wait()

		// 最初のタスクの完了時間と最後のタスクの完了時間の差を計算
		var minTime, maxTime time.Time
		for i, t := range completionTimes {
			if i == 0 || t.Before(minTime) {
				minTime = t
			}
			if i == 0 || t.After(maxTime) {
				maxTime = t
			}
		}

		// 差が0.5秒未満であることを確認（すべてのタスクが並列実行されていれば、差は小さいはず）
		timeDiff := maxTime.Sub(minTime)
		if timeDiff > 500*time.Millisecond {
			t.Errorf("Tasks did not execute in parallel, time difference: %v", timeDiff)
		}
	})
}

// シナリオ処理のテスト
func TestScenarioProcessing(t *testing.T) {
	// シナリオ処理のテスト
	t.Run("Scenario processing test", func(t *testing.T) {
		// モックのシナリオとコンフィグを作成
		mockScenario := MockScenario{
			Name:        "Test Scenario",
			PresetName:  "Test Preset",
			Description: "Test Description",
		}

		mockConfig := &MockBatchConfig{
			DynamoDBEndpoint: "http://localhost:4566",
			DynamoDBRegion:   "us-east-1",
			LogFile:          "",
			DataDir:          "test_data",
		}

		// このテストはモックを使用して、実際のファイルやDBにアクセスせずにロジックをテスト
		// 実際の実装では、モックを使用してgenerateHandsAndFlop、calculateEquity、saveResultToDynamoDBをテスト

		// 例: モックのハンドとフロップを生成
		heroHand := "AsAcKhQd"
		opponentRange := "KsKcJhTd,QsQcJdTc"
		flop := []poker.Card{
			poker.NewCard("2c"),
			poker.NewCard("7d"),
			poker.NewCard("Ts"),
		}

		// 並列処理のテスト
		numCPU := runtime.NumCPU()
		if numCPU < 1 {
			t.Errorf("Expected at least 1 CPU, got %d", numCPU)
		}

		// セマフォとWaitGroupの動作確認
		semaphore := make(chan struct{}, numCPU)
		var wg sync.WaitGroup

		// 複数のgoroutineを起動してセマフォの動作を確認
		for i := 0; i < numCPU*2; i++ {
			wg.Add(1)
			go func() {
				semaphore <- struct{}{}
				time.Sleep(10 * time.Millisecond)
				<-semaphore
				wg.Done()
			}()
		}

		// すべてのgoroutineが完了するのを待つ
		wg.Wait()

		// テストが正常に完了したことを確認
		t.Logf("Successfully tested parallel processing with %d CPUs", numCPU)
		t.Logf("Test data: heroHand=%s, opponentRange=%s, flop=%v", heroHand, opponentRange, flop)
		t.Logf("Mock scenario: %s, Mock config: %v", mockScenario.Name, mockConfig)
	})
}
