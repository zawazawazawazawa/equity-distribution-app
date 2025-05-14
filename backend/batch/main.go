package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chehsunliu/poker"
	"github.com/joho/godotenv"

	"equity-distribution-backend/pkg/db"
	"equity-distribution-backend/pkg/fileio"
	pkrlib "equity-distribution-backend/pkg/poker"
)

// シナリオ定義
type Scenario struct {
	Name           string
	PresetName     string
	Description    string
	HeroHandRanges []string // ヒーローハンドの範囲（将来的な拡張用）
}

// 利用可能なシナリオのリスト
var scenarios = []Scenario{
	{
		Name:        "SRP BB vs UTG",
		PresetName:  "SRP BB call vs UTG open",
		Description: "シングルレイズポット: BBがUTGオープンに対してコール",
	},
	{
		Name:        "SRP BB vs BTN",
		PresetName:  "SRP BB call vs BTN open",
		Description: "シングルレイズポット: BBがBTNオープンに対してコール",
	},
	{
		Name:        "SRP BTN vs UTG",
		PresetName:  "SRP BTN call vs UTG open",
		Description: "シングルレイズポット: BTNがUTGオープンに対してコール",
	},
	{
		Name:        "3BP UTG vs BB",
		PresetName:  "3BP UTG call vs BB 3bet",
		Description: "3ベットポット: UTGがBBの3ベットに対してコール",
	},
	{
		Name:        "3BP UTG vs BTN",
		PresetName:  "3BP UTG call vs BTN 3bet",
		Description: "3ベットポット: UTGがBTNの3ベットに対してコール",
	},
	{
		Name:        "3BP BTN vs BB",
		PresetName:  "3BP BTN call vs BB 3bet",
		Description: "3ベットポット: BTNがBBの3ベットに対してコール",
	},
}

// EquityResult は1つのシナリオの計算結果を表します
type EquityResult struct {
	Scenario      Scenario
	HeroHand      string
	Flop          []poker.Card
	Equities      map[string]float64
	AverageEquity float64 // 平均エクイティ
}

// バッチ処理の設定
type BatchConfig struct {
	LogFile string // ログファイル
	DataDir string // データディレクトリ

	// PostgreSQL設定
	PostgresHost     string // PostgreSQLホスト
	PostgresPort     int    // PostgreSQLポート
	PostgresUser     string // PostgreSQLユーザー
	PostgresPassword string // PostgreSQLパスワード
	PostgresDBName   string // PostgreSQLデータベース名
}

func main() {
	// .envファイルの読み込み
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// コマンドライン引数の解析
	config := parseFlags()

	// ログの設定
	setupLogging(config.LogFile)

	// 乱数生成器の初期化
	rand.Seed(time.Now().UnixNano())

	// PostgreSQL接続の確立
	pgConfig := db.PostgresConfig{
		Host:     config.PostgresHost,
		Port:     config.PostgresPort,
		User:     config.PostgresUser,
		Password: config.PostgresPassword,
		DBName:   config.PostgresDBName,
	}

	pgDB, err := db.GetPostgresConnection(pgConfig)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgDB.Close()

	// 並列処理の設定
	numCPU := runtime.NumCPU()
	log.Printf("Starting parallel processing for %d scenarios using %d CPUs", len(scenarios), numCPU)

	// 同時実行数を制限するセマフォ
	semaphore := make(chan struct{}, numCPU)
	var wg sync.WaitGroup

	// 結果を収集するためのチャネル
	resultChan := make(chan EquityResult, len(scenarios))

	// 各シナリオを並列で実行
	for i, scenario := range scenarios {
		wg.Add(1)
		semaphore <- struct{}{} // セマフォを取得

		// シナリオ処理をgoroutineで実行
		go func(index int, currentScenario Scenario) {
			defer wg.Done()
			defer func() { <-semaphore }() // セマフォを解放

			log.Printf("Starting scenario %d: %s", index+1, currentScenario.Name)
			log.Printf("Selected scenario: %s", currentScenario.Name)

			// シナリオに基づいてハンドとフロップを生成
			heroHand, opponentRange, flop := generateHandsAndFlop(currentScenario, config)

			// equity計算
			equities, err := calculateEquity(heroHand, opponentRange, flop, config)
			if err != nil {
				log.Printf("Error calculating equity: %v", err)
				log.Printf("Scenario %d failed: %v", index+1, err)
				return
			}

			// 平均エクイティの計算
			var totalEquity float64
			for _, equity := range equities {
				totalEquity += equity
			}
			averageEquity := totalEquity / float64(len(equities))

			// 結果をチャネルに送信
			resultChan <- EquityResult{
				Scenario:      currentScenario,
				HeroHand:      heroHand,
				Flop:          flop,
				Equities:      equities,
				AverageEquity: averageEquity,
			}

			log.Printf("Scenario %d completed: %s - Flop: %s, Hero: %s, Average Equity: %.2f%%",
				index+1, currentScenario.Name, pkrlib.GenerateBoardString(flop), heroHand, averageEquity)
		}(i, scenario)
	}

	// すべてのgoroutineが完了するのを待つ
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 結果を収集
	var results []EquityResult
	for result := range resultChan {
		results = append(results, result)
	}

	// 翌日の日付を取得
	tomorrow := time.Now().AddDate(0, 0, 1)

	// 結果をシナリオごとにグループ化
	scenarioResults := make(map[string][]EquityResult)
	for _, result := range results {
		scenarioName := result.Scenario.Name
		scenarioResults[scenarioName] = append(scenarioResults[scenarioName], result)
	}

	// 各シナリオごとに一つのレコードとして保存
	for scenarioName, scenarioResultList := range scenarioResults {
		if len(scenarioResultList) == 0 {
			continue
		}

		// VillainEquity構造体を定義
		type VillainEquity struct {
			VillainHand string  `json:"villain_hand"`
			Equity      float64 `json:"equity"`
		}

		// このシナリオのすべての結果から対戦相手のハンドとエクイティの配列を作成
		allVillainEquities := []VillainEquity{}

		// 平均エクイティの計算用
		totalEquity := 0.0

		// 最初の結果からheroHandとflopを取得（代表値として）
		var heroHand string
		var flop string
		if len(scenarioResultList) > 0 {
			heroHand = scenarioResultList[0].HeroHand
			flop = pkrlib.GenerateBoardString(scenarioResultList[0].Flop)
		}

		for _, result := range scenarioResultList {
			// Equitiesマップを配列に変換して追加
			for villainHand, equity := range result.Equities {
				allVillainEquities = append(allVillainEquities, VillainEquity{
					VillainHand: villainHand,
					Equity:      equity,
				})
			}

			totalEquity += result.AverageEquity
		}

		// VillainEquities配列をJSON文字列に変換
		villainEquitiesJSON, err := json.Marshal(allVillainEquities)
		if err != nil {
			log.Printf("Error marshaling villain equities for scenario %s to JSON: %v", scenarioName, err)
			continue
		}

		// 平均エクイティはすべての結果の平均を使用
		averageEquity := totalEquity / float64(len(scenarioResultList))

		// シナリオごとに一つのレコードとしてPostgreSQLに保存
		err = db.InsertDailyQuizResult(
			pgDB,
			tomorrow,
			scenarioName,
			heroHand,
			flop,
			string(villainEquitiesJSON), // VillainEquitiesの配列だけをJSON文字列として保存
			averageEquity,
		)
		if err != nil {
			log.Printf("Error saving villain equities for scenario %s to PostgreSQL: %v", scenarioName, err)
		} else {
			log.Printf("Successfully saved villain equities for scenario %s as one record to PostgreSQL", scenarioName)
		}
	}

	log.Println("All scenarios processed successfully")
}

// コマンドライン引数を解析する
func parseFlags() *BatchConfig {
	config := &BatchConfig{}

	flag.StringVar(&config.LogFile, "log", "", "Log file (empty for stdout)")
	flag.StringVar(&config.DataDir, "data", "data", "Directory containing preset data files")

	// PostgreSQL設定
	flag.StringVar(&config.PostgresHost, "pg-host", "localhost", "PostgreSQL host")
	flag.IntVar(&config.PostgresPort, "pg-port", 5432, "PostgreSQL port")
	flag.StringVar(&config.PostgresUser, "pg-user", "postgres", "PostgreSQL user")
	flag.StringVar(&config.PostgresPassword, "pg-password", "postgres", "PostgreSQL password")
	flag.StringVar(&config.PostgresDBName, "pg-dbname", "plo_equity", "PostgreSQL database name")

	flag.Parse()

	return config
}

// ログの設定
func setupLogging(logFile string) {
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		log.SetOutput(f)
	}
}

// シナリオに基づいてハンドとフロップを生成する
func generateHandsAndFlop(scenario Scenario, config *BatchConfig) (string, string, []poker.Card) {
	// Opponentレンジはプリセットから読み込む
	opponentRange, err := fileio.LoadOpponentRangeFromPreset(scenario.PresetName, config.DataDir)
	if err != nil {
		// 失敗したらpanicを投げる
		panic(fmt.Sprintf("Failed to load opponent range: %v", err))
	}

	// アグレッサー側のレンジを読み込む
	aggressorRange, err := fileio.LoadAggressorRangeFromPreset(scenario.PresetName, config.DataDir)
	if err != nil {
		log.Printf("Error loading aggressor range: %v", err)
		// 失敗したらpanicを投げる
		panic(fmt.Sprintf("Failed to load aggressor range: %v", err))
	}

	// アグレッサー側のレンジからランダムに1ハンドを選ぶ
	aggressorHands := strings.Split(aggressorRange, ",")
	if len(aggressorHands) == 0 {
		// 失敗したらpanicを投げる
		panic("No aggressor hands found")
	}
	heroHand := aggressorHands[rand.Intn(len(aggressorHands))]
	log.Printf("Selected hero hand from aggressor range: %s", heroHand)

	// heroHandに含まれるカードは除外して、flopをランダムに生成
	// ヒーローハンドをpoker.Card形式に変換
	var heroCards []poker.Card
	if len(heroHand) == 8 { // PLOハンド（4枚）
		for j := 0; j < 8; j += 2 {
			cardStr := strings.ToUpper(heroHand[j:j+1]) + strings.ToLower(heroHand[j+1:j+2])
			card := poker.NewCard(cardStr)
			heroCards = append(heroCards, card)
		}
	} else {
		log.Printf("Warning: Unexpected hero hand format: %s", heroHand)
	}

	// 52枚のデッキを生成
	deck := poker.NewDeck()
	fullDeck := deck.Draw(52) // 52枚すべてのカードを取得

	// ヒーローハンドに含まれるカードを除外
	remainingDeck := []poker.Card{}
	for _, card := range fullDeck {
		isInHeroHand := false
		for _, heroCard := range heroCards {
			if card.String() == heroCard.String() {
				isInHeroHand = true
				break
			}
		}
		if !isInHeroHand {
			remainingDeck = append(remainingDeck, card)
		}
	}

	// 残りのカードからランダムに3枚選んでフロップとする
	flop := []poker.Card{}
	for i := 0; i < 3; i++ {
		if len(remainingDeck) == 0 {
			break
		}
		idx := rand.Intn(len(remainingDeck))
		flop = append(flop, remainingDeck[idx])
		// 選んだカードを削除（重複を避けるため）
		remainingDeck = append(remainingDeck[:idx], remainingDeck[idx+1:]...)
	}

	log.Printf("Generated flop: %s", pkrlib.GenerateBoardString(flop))
	return heroHand, opponentRange, flop
}

// equity計算を実行する
func calculateEquity(heroHand string, opponentRange string, flop []poker.Card, config *BatchConfig) (map[string]float64, error) {
	// ヒーローハンドをpoker.Card形式に変換
	var yourHand []poker.Card
	if len(heroHand) == 8 {
		for j := 0; j < 8; j += 2 {
			cardStr := strings.ToUpper(heroHand[j:j+1]) + strings.ToLower(heroHand[j+1:j+2])
			tempCard := poker.NewCard(cardStr)
			yourHand = append(yourHand, tempCard)
		}
	} else {
		return nil, fmt.Errorf("invalid hero hand format: %s", heroHand)
	}

	// Opponentレンジをpoker.Card形式に変換
	opponentHands := strings.Split(opponentRange, ",")
	var formattedOpponentHands [][]poker.Card
	for _, hand := range opponentHands {
		tmpHand := strings.Split(hand, "@")[0]
		var tempArray []poker.Card
		if len(tmpHand) == 8 {
			for j := 0; j < 8; j += 2 {
				cardStr := strings.ToUpper(tmpHand[j:j+1]) + strings.ToLower(tmpHand[j+1:j+2])
				tempCard := poker.NewCard(cardStr)
				tempArray = append(tempArray, tempCard)
			}
			formattedOpponentHands = append(formattedOpponentHands, tempArray)
		}
	}

	// 共通の並列計算関数を使用してequity計算を実行
	return pkrlib.CalculateHandVsRangeEquityParallel(yourHand, formattedOpponentHands, flop)
}
