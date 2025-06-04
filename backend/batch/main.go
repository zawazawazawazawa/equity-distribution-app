package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chehsunliu/poker"
	"github.com/joho/godotenv"

	"equity-distribution-backend/pkg/db"
	"equity-distribution-backend/pkg/fileio"
	"equity-distribution-backend/pkg/image"
	pkrlib "equity-distribution-backend/pkg/poker"
	"equity-distribution-backend/pkg/storage"
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
		Name:        "SRP UTG vs BB",
		PresetName:  "SRP BB call vs UTG open",
		Description: "シングルレイズポット: BBがUTGオープンに対してコール",
	},
	{
		Name:        "SRP BTN vs BB",
		PresetName:  "SRP BB call vs BTN open",
		Description: "シングルレイズポット: BBがBTNオープンに対してコール",
	},
	{
		Name:        "SRP UTG vs BTN",
		PresetName:  "SRP BTN call vs UTG open",
		Description: "シングルレイズポット: BTNがUTGオープンに対してコール",
	},
	{
		Name:        "3BP BB vs UTG",
		PresetName:  "3BP UTG call vs BB 3bet",
		Description: "3ベットポット: UTGがBBの3ベットに対してコール",
	},
	{
		Name:        "3BP BTN vs UTG",
		PresetName:  "3BP UTG call vs BTN 3bet",
		Description: "3ベットポット: UTGがBTNの3ベットに対してコール",
	},
	{
		Name:        "3BP BB vs BTN",
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
	Date    string // 日付（YYYY-MM-DD形式）

	// PostgreSQL設定
	PostgresHost     string // PostgreSQLホスト
	PostgresPort     int    // PostgreSQLポート
	PostgresUser     string // PostgreSQLユーザー
	PostgresPassword string // PostgreSQLパスワード
	PostgresDBName   string // PostgreSQLデータベース名

	// 並列処理の設定
	EnableParallelProcessing bool // 並列処理の有効/無効
	MaxParallelJobs          int  // 最大同時実行数

	// 画像アップロード設定
	EnableImageUpload bool // 画像アップロードの有効/無効
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
	var results []EquityResult

	// 日付の処理
	var targetDate time.Time
	if config.Date == "" {
		// 日付が指定されていない場合は翌日の日付を使用
		targetDate = time.Now().AddDate(0, 0, 1)
	} else {
		// 指定された日付を解析
		var err error
		targetDate, err = time.Parse("2006-01-02", config.Date)
		if err != nil {
			log.Fatalf("Invalid date format: %v. Please use YYYY-MM-DD format.", err)
		}
	}

	// 指定された日付のデータがデータベースに既に存在するか確認
	existingResults, err := db.GetDailyQuizResultsByDate(pgDB, targetDate)
	if err != nil {
		log.Printf("Error checking existing data: %v", err)
	}

	// existingResultsが空でない場合は、すでにデータが存在するため、existingResultsをresultsに変換して代入
	if len(existingResults) > 0 {
		log.Printf("Data for %s already exists in the database. Skipping processing.", targetDate.Format("2006-01-02"))

		// []map[string]interface{}から[]EquityResultに変換
		for _, result := range existingResults {
			scenarioName, _ := result["scenario"].(string)
			heroHand, _ := result["hero_hand"].(string)
			flopStr, _ := result["flop"].(string) // フロップ文字列を取得
			averageEquity, _ := result["average_equity"].(float64)

			// シナリオの検索
			var foundScenario Scenario
			for _, s := range scenarios {
				if s.Name == scenarioName {
					foundScenario = s
					break
				}
			}

			// フロップの文字列をpoker.Card配列に変換
			var flopCards []poker.Card
			if flopStr != "" {
				// フロップ文字列は "2d3cJc" のような形式と仮定
				// 2文字ずつ（数字+スート）で分割して処理
				for i := 0; i < len(flopStr); i += 2 {
					if i+2 <= len(flopStr) {
						cardStr := strings.ToUpper(flopStr[i:i+1]) + strings.ToLower(flopStr[i+1:i+2])
						card := poker.NewCard(cardStr)
						flopCards = append(flopCards, card)
					}
				}
				log.Printf("Using flop cards: %s", pkrlib.GenerateBoardString(flopCards))
			} else {
				log.Printf("Warning: No flop data found for date %s", targetDate.Format("2006-01-02"))
			}

			// Equitiesマップの作成（データベースから取得したデータに基づく）
			equities := make(map[string]float64)

			results = append(results, EquityResult{
				Scenario:      foundScenario,
				HeroHand:      heroHand,
				Flop:          flopCards,
				Equities:      equities,
				AverageEquity: averageEquity,
			})
		}
	} else {
		// 計算処理に進む
		if config.EnableParallelProcessing {
			// 並列処理が有効な場合
			maxJobs := config.MaxParallelJobs
			log.Printf("Starting parallel processing for %d scenarios using %d jobs", len(scenarios), maxJobs)

			// 同時実行数を制限するセマフォ
			semaphore := make(chan struct{}, maxJobs)
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
			for result := range resultChan {
				results = append(results, result)
			}
		} else {
			// 並列処理が無効な場合（シーケンシャル処理）
			log.Printf("Starting sequential processing for %d scenarios", len(scenarios))

			// 各シナリオを順次実行
			for i, scenario := range scenarios {
				log.Printf("Starting scenario %d: %s", i+1, scenario.Name)
				log.Printf("Selected scenario: %s", scenario.Name)

				// シナリオに基づいてハンドとフロップを生成
				heroHand, opponentRange, flop := generateHandsAndFlop(scenario, config)

				// equity計算
				equities, err := calculateEquity(heroHand, opponentRange, flop, config)
				if err != nil {
					log.Printf("Error calculating equity: %v", err)
					log.Printf("Scenario %d failed: %v", i+1, err)
					continue
				}

				// 平均エクイティの計算
				var totalEquity float64
				for _, equity := range equities {
					totalEquity += equity
				}
				averageEquity := totalEquity / float64(len(equities))

				// 結果を追加
				results = append(results, EquityResult{
					Scenario:      scenario,
					HeroHand:      heroHand,
					Flop:          flop,
					Equities:      equities,
					AverageEquity: averageEquity,
				})

				log.Printf("Scenario %d completed: %s - Flop: %s, Hero: %s, Average Equity: %.2f%%",
					i+1, scenario.Name, pkrlib.GenerateBoardString(flop), heroHand, averageEquity)
			}
		}

		// バッチ処理用のデータを準備
		var batchResults []db.DailyQuizResult

		// 結果をシナリオごとにグループ化
		scenarioResults := make(map[string][]EquityResult)
		for _, result := range results {
			scenarioName := result.Scenario.Name
			scenarioResults[scenarioName] = append(scenarioResults[scenarioName], result)
		}

		// 各シナリオごとにバッチ用データを作成
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

			// バッチ用データに追加
			batchResults = append(batchResults, db.DailyQuizResult{
				Date:          targetDate,
				Scenario:      scenarioName,
				HeroHand:      heroHand,
				Flop:          flop,
				Result:        string(villainEquitiesJSON),
				AverageEquity: averageEquity,
				GameType:      "4card_plo", // ゲームタイプを明示的に指定
			})
		}

		// バッチ処理でPostgreSQLに保存
		if len(batchResults) > 0 {
			log.Printf("Starting batch insert of %d records to PostgreSQL", len(batchResults))
			err = db.InsertDailyQuizResultsBatch(pgDB, batchResults)
			if err != nil {
				log.Printf("Error in batch insert to PostgreSQL: %v", err)
			} else {
				log.Printf("Successfully completed batch insert to PostgreSQL")
			}
		}

		log.Println("All scenarios processed successfully")
	}

	// 画像アップロードが有効な場合のみ画像生成とアップロードを実行
	if config.EnableImageUpload {
		log.Println("Image upload is enabled. Starting image generation and upload...")

		// 1問目のシナリオを使用して画像生成
		if len(results) > 0 {
			firstResult := results[0]
			imagePath := filepath.Join("images/daily-quiz", targetDate.Format("2006-01-02")+".png")

			err := image.GenerateDailyQuizImage(
				targetDate,
				firstResult.Scenario.Name,
				firstResult.HeroHand,
				firstResult.Flop,
			)
			if err != nil {
				log.Printf("Error generating daily quiz image: %v", err)
			} else {
				log.Printf("Successfully generated daily quiz image for %s", targetDate.Format("2006-01-02"))

				// R2設定を取得
				r2Config := storage.R2Config{
					Endpoint:   getEnvOrDefault("R2_ENDPOINT", ""),
					AccessKey:  getEnvOrDefault("R2_ACCESS_KEY", ""),
					SecretKey:  getEnvOrDefault("R2_SECRET_KEY", ""),
					BucketName: getEnvOrDefault("R2_BUCKET", ""),
				}

				// R2クライアントを作成
				r2Client, err := storage.GetR2Client(r2Config)
				if err != nil {
					log.Printf("Error creating R2 client: %v", err)
				} else {
					// 画像をR2にアップロード
					objectKey := "daily-quiz/" + targetDate.Format("2006-01-02") + ".png"
					err = storage.UploadImageToR2(r2Client, r2Config.BucketName, imagePath, objectKey)
					if err != nil {
						log.Printf("Error uploading image to R2: %v", err)
					} else {
						log.Printf("Successfully uploaded image to R2: %s", objectKey)

						// 公開URLを生成
						publicURL := storage.GetR2ObjectURL(r2Config.Endpoint, r2Config.BucketName, objectKey)
						log.Printf("Image public URL: %s", publicURL)
					}
				}
			}

			// 明示的にGCを呼び出し
			runtime.GC()

			log.Println("Image generation and upload completed")
		}
	} else {
		log.Println("Image upload is disabled. Skipping image generation and upload.")
	}
}

// コマンドライン引数を解析する
func parseFlags() *BatchConfig {
	config := &BatchConfig{}

	// 環境変数からデフォルト値を取得
	postgresHost := getEnvOrDefault("POSTGRES_HOST", "localhost")
	postgresPort := getEnvIntOrDefault("POSTGRES_PORT", 5432)
	postgresUser := getEnvOrDefault("POSTGRES_USER", "postgres")
	postgresPassword := getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
	postgresDBName := getEnvOrDefault("POSTGRES_DBNAME", "plo_equity")

	// 並列処理の設定を環境変数から取得
	enableParallelProcessing := getEnvBoolOrDefault("ENABLE_PARALLEL_PROCESSING", true)
	maxParallelJobs := getEnvIntOrDefault("MAX_PARALLEL_JOBS", runtime.NumCPU())

	// 画像アップロード設定を環境変数から取得
	enableImageUpload := getEnvBoolOrDefault("ENABLE_IMAGE_UPLOAD", true)

	flag.StringVar(&config.LogFile, "log", "", "Log file (empty for stdout)")
	flag.StringVar(&config.DataDir, "data", "data", "Directory containing preset data files")
	flag.StringVar(&config.Date, "date", "", "Date for quiz in YYYY-MM-DD format (default: tomorrow)")

	// PostgreSQL設定（環境変数のデフォルト値を使用）
	flag.StringVar(&config.PostgresHost, "pg-host", postgresHost, "PostgreSQL host")
	flag.IntVar(&config.PostgresPort, "pg-port", postgresPort, "PostgreSQL port")
	flag.StringVar(&config.PostgresUser, "pg-user", postgresUser, "PostgreSQL user")
	flag.StringVar(&config.PostgresPassword, "pg-password", postgresPassword, "PostgreSQL password")
	flag.StringVar(&config.PostgresDBName, "pg-dbname", postgresDBName, "PostgreSQL database name")

	// 並列処理の設定
	flag.BoolVar(&config.EnableParallelProcessing, "parallel", enableParallelProcessing, "Enable parallel processing")
	flag.IntVar(&config.MaxParallelJobs, "jobs", maxParallelJobs, "Maximum number of parallel jobs")

	// 画像アップロード設定
	flag.BoolVar(&config.EnableImageUpload, "image-upload", enableImageUpload, "Enable image generation and upload")

	flag.Parse()

	return config
}

// 環境変数から文字列値を取得するヘルパー関数
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// 環境変数から整数値を取得するヘルパー関数
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 環境変数からブール値を取得するヘルパー関数
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
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

	// 並列処理の設定に応じてequity計算を実行
	if config.EnableParallelProcessing {
		// 並列処理が有効な場合は並列計算関数を使用
		return pkrlib.CalculateHandVsRangeEquityParallel(yourHand, formattedOpponentHands, flop)
	} else {
		// 並列処理が無効な場合は非並列計算関数を使用
		// 注: pkrlib.CalculateHandVsRangeEquityという非並列版の関数が存在しない場合は、
		// 並列版の関数を使用します
		return pkrlib.CalculateHandVsRangeEquityParallel(yourHand, formattedOpponentHands, flop)
	}
}
