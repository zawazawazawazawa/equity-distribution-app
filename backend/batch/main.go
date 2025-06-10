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
	{
		Name:        "PLO5 SRP UTG vs BB",
		PresetName:  "PLO5 SRP BB call vs UTG open",
		Description: "5-card PLO シングルレイズポット: BBがUTGオープンに対してコール",
	},
	{
		Name:        "PLO5 SRP BTN vs BB",
		PresetName:  "PLO5 SRP BB call vs BTN open",
		Description: "5-card PLO シングルレイズポット: BBがBTNオープンに対してコール",
	},
	{
		Name:        "PLO5 SRP UTG vs BTN",
		PresetName:  "PLO5 SRP BTN call vs UTG open",
		Description: "5-card PLO シングルレイズポット: BTNがUTGオープンに対してコール",
	},
	{
		Name:        "PLO5 3BP BB vs UTG",
		PresetName:  "PLO5 3BP UTG call vs BB 3bet",
		Description: "5-card PLO 3ベットポット: UTGがBBの3ベットに対してコール",
	},
	{
		Name:        "PLO5 3BP BTN vs UTG",
		PresetName:  "PLO5 3BP UTG call vs BTN 3bet",
		Description: "5-card PLO 3ベットポット: UTGがBTNの3ベットに対してコール",
	},
	{
		Name:        "PLO5 3BP BB vs BTN",
		PresetName:  "PLO5 3BP BTN call vs BB 3bet",
		Description: "5-card PLO 3ベットポット: BTNがBBの3ベットに対してコール",
	},
}

// EquityResult は1つのシナリオの計算結果を表します
type EquityResult struct {
	Scenario      Scenario
	HeroHand      string
	Flop          []poker.Card
	Equities      map[string]float64
	AverageEquity float64 // 平均エクイティ
	SamplingCount int     // Adaptive samplingで使用したサンプル数
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

	// エクイティ計算設定
	UseMonteCarloEquity bool   // Monte Carlo法を使用するか（false: exhaustive）
	MonteCarloMode      string // Monte Carloの精度モード（FAST/NORMAL/ACCURATE）
	UseAdaptiveSampling bool   // Adaptive samplingを使用するか
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
					equities, samplingCount, err := calculateEquity(heroHand, opponentRange, flop, config)
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
						SamplingCount: samplingCount,
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
				equities, samplingCount, err := calculateEquity(heroHand, opponentRange, flop, config)
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
					SamplingCount: samplingCount,
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
			var totalSamplingCount int

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
				totalSamplingCount += result.SamplingCount
			}

			// VillainEquities配列をJSON文字列に変換
			villainEquitiesJSON, err := json.Marshal(allVillainEquities)
			if err != nil {
				log.Printf("Error marshaling villain equities for scenario %s to JSON: %v", scenarioName, err)
				continue
			}

			// 平均エクイティはすべての結果の平均を使用
			averageEquity := totalEquity / float64(len(scenarioResultList))

			// ゲームタイプの判定
			gameType := "4card_plo"
			if len(heroHand) == 10 {
				gameType = "5card_plo"
			}

			// サンプリング数の処理（0の場合はnilを設定）
			var samplingCountPtr *int
			if totalSamplingCount > 0 {
				samplingCountPtr = &totalSamplingCount
			}

			// バッチ用データに追加
			batchResults = append(batchResults, db.DailyQuizResult{
				Date:          targetDate,
				Scenario:      scenarioName,
				HeroHand:      heroHand,
				Flop:          flop,
				Result:        string(villainEquitiesJSON),
				AverageEquity: averageEquity,
				GameType:      gameType,
				SamplingCount: samplingCountPtr,
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

		// 4-card PLOと5-card PLOそれぞれから1問ずつ選択
		var fourCardResult *EquityResult
		var fiveCardResult *EquityResult

		// resultsから4-cardと5-cardの問題を抽出
		for i := range results {
			if len(results[i].HeroHand) == 8 && fourCardResult == nil {
				fourCardResult = &results[i]
			} else if len(results[i].HeroHand) == 10 && fiveCardResult == nil {
				fiveCardResult = &results[i]
			}
			
			// 両方見つかったら終了
			if fourCardResult != nil && fiveCardResult != nil {
				break
			}
		}

		// R2設定を取得（両方の画像で共通使用）
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
		}

		// 4-card PLOの画像生成とアップロード
		if fourCardResult != nil {
			gameTypeDir := "4card"
			imagePath := filepath.Join("images/daily-quiz", gameTypeDir, targetDate.Format("2006-01-02")+".png")

			err := image.GenerateDailyQuizImage(
				targetDate,
				fourCardResult.Scenario.Name,
				fourCardResult.HeroHand,
				fourCardResult.Flop,
			)
			if err != nil {
				log.Printf("Error generating 4-card PLO daily quiz image: %v", err)
			} else {
				log.Printf("Successfully generated 4-card PLO daily quiz image for %s", targetDate.Format("2006-01-02"))

				if r2Client != nil {
					// 画像をR2にアップロード
					objectKey := "daily-quiz/" + gameTypeDir + "/" + targetDate.Format("2006-01-02") + ".png"
					err = storage.UploadImageToR2(r2Client, r2Config.BucketName, imagePath, objectKey)
					if err != nil {
						log.Printf("Error uploading 4-card PLO image to R2: %v", err)
					} else {
						log.Printf("Successfully uploaded 4-card PLO image to R2: %s", objectKey)

						// 公開URLを生成
						publicURL := storage.GetR2ObjectURL(r2Config.Endpoint, r2Config.BucketName, objectKey)
						log.Printf("4-card PLO image public URL: %s", publicURL)
					}
				}
			}
		} else {
			log.Printf("No 4-card PLO result found for image generation")
		}

		// 5-card PLOの画像生成とアップロード
		if fiveCardResult != nil {
			gameTypeDir := "5card"
			imagePath := filepath.Join("images/daily-quiz", gameTypeDir, targetDate.Format("2006-01-02")+".png")

			err := image.GenerateDailyQuizImage(
				targetDate,
				fiveCardResult.Scenario.Name,
				fiveCardResult.HeroHand,
				fiveCardResult.Flop,
			)
			if err != nil {
				log.Printf("Error generating 5-card PLO daily quiz image: %v", err)
			} else {
				log.Printf("Successfully generated 5-card PLO daily quiz image for %s", targetDate.Format("2006-01-02"))

				if r2Client != nil {
					// 画像をR2にアップロード
					objectKey := "daily-quiz/" + gameTypeDir + "/" + targetDate.Format("2006-01-02") + ".png"
					err = storage.UploadImageToR2(r2Client, r2Config.BucketName, imagePath, objectKey)
					if err != nil {
						log.Printf("Error uploading 5-card PLO image to R2: %v", err)
					} else {
						log.Printf("Successfully uploaded 5-card PLO image to R2: %s", objectKey)

						// 公開URLを生成
						publicURL := storage.GetR2ObjectURL(r2Config.Endpoint, r2Config.BucketName, objectKey)
						log.Printf("5-card PLO image public URL: %s", publicURL)
					}
				}
			}
		} else {
			log.Printf("No 5-card PLO result found for image generation")
		}

		// 明示的にGCを呼び出し
		runtime.GC()

		log.Println("Image generation and upload completed")
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

	// エクイティ計算設定を環境変数から取得
	useMonteCarloEquity := getEnvBoolOrDefault("USE_MONTE_CARLO_EQUITY", false)
	monteCarloMode := getEnvOrDefault("MONTE_CARLO_MODE", "ACCURATE")
	useAdaptiveSampling := getEnvBoolOrDefault("USE_ADAPTIVE_SAMPLING", false)

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

	// エクイティ計算設定
	flag.BoolVar(&config.UseMonteCarloEquity, "monte-carlo", useMonteCarloEquity, "Use Monte Carlo equity calculation instead of exhaustive")
	flag.StringVar(&config.MonteCarloMode, "monte-carlo-mode", monteCarloMode, "Monte Carlo accuracy mode (FAST/NORMAL/ACCURATE)")
	flag.BoolVar(&config.UseAdaptiveSampling, "adaptive", useAdaptiveSampling, "Use adaptive sampling for hand vs range calculation")

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
	if len(heroHand) == 8 { // 4-card PLOハンド（4枚）
		for j := 0; j < 8; j += 2 {
			cardStr := strings.ToUpper(heroHand[j:j+1]) + strings.ToLower(heroHand[j+1:j+2])
			card := poker.NewCard(cardStr)
			heroCards = append(heroCards, card)
		}
	} else if len(heroHand) == 10 { // 5-card PLOハンド（5枚）
		for j := 0; j < 10; j += 2 {
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
func calculateEquity(heroHand string, opponentRange string, flop []poker.Card, config *BatchConfig) (map[string]float64, int, error) {
	// ヒーローハンドをpoker.Card形式に変換
	var yourHand []poker.Card
	if len(heroHand) == 8 { // 4-card PLO
		for j := 0; j < 8; j += 2 {
			cardStr := strings.ToUpper(heroHand[j:j+1]) + strings.ToLower(heroHand[j+1:j+2])
			tempCard := poker.NewCard(cardStr)
			yourHand = append(yourHand, tempCard)
		}
	} else if len(heroHand) == 10 { // 5-card PLO
		for j := 0; j < 10; j += 2 {
			cardStr := strings.ToUpper(heroHand[j:j+1]) + strings.ToLower(heroHand[j+1:j+2])
			tempCard := poker.NewCard(cardStr)
			yourHand = append(yourHand, tempCard)
		}
	} else {
		return nil, 0, fmt.Errorf("invalid hero hand format: %s", heroHand)
	}

	// Opponentレンジをpoker.Card形式に変換
	opponentHands := strings.Split(opponentRange, ",")
	var formattedOpponentHands [][]poker.Card
	for _, hand := range opponentHands {
		tmpHand := strings.Split(hand, "@")[0]
		var tempArray []poker.Card
		if len(tmpHand) == 8 { // 4-card PLO
			for j := 0; j < 8; j += 2 {
				cardStr := strings.ToUpper(tmpHand[j:j+1]) + strings.ToLower(tmpHand[j+1:j+2])
				tempCard := poker.NewCard(cardStr)
				tempArray = append(tempArray, tempCard)
			}
			formattedOpponentHands = append(formattedOpponentHands, tempArray)
		} else if len(tmpHand) == 10 { // 5-card PLO
			for j := 0; j < 10; j += 2 {
				cardStr := strings.ToUpper(tmpHand[j:j+1]) + strings.ToLower(tmpHand[j+1:j+2])
				tempCard := poker.NewCard(cardStr)
				tempArray = append(tempArray, tempCard)
			}
			formattedOpponentHands = append(formattedOpponentHands, tempArray)
		}
	}

	// Adaptive Samplingを使用するかどうかで分岐
	if config.UseAdaptiveSampling {
		// Adaptive Sampling法を使用（レンジ全体を動的サンプリング）
		log.Printf("Using adaptive sampling for hand vs range calculation")
		
		// Adaptive設定を作成
		adaptiveConfig := pkrlib.DefaultAdaptiveConfig()
		switch config.MonteCarloMode {
		case "FAST":
			adaptiveConfig.MinSamples = 500
			adaptiveConfig.MaxSamples = 3000
			adaptiveConfig.TargetError = 0.03 // ±3%誤差
		case "ACCURATE":
			adaptiveConfig.MinSamples = 2000
			adaptiveConfig.MaxSamples = 15000
			adaptiveConfig.TargetError = 0.005 // ±0.5%誤差
		case "NORMAL":
			// デフォルト設定をそのまま使用
		}
		
		// Adaptive samplingで計算（個別のエクイティも取得）
		equities, avgEquity, samplesUsed, err := pkrlib.CalculateHandVsRangeAdaptiveWithDetails(
			yourHand, formattedOpponentHands, flop, adaptiveConfig,
		)
		if err != nil {
			return nil, 0, err
		}
		
		log.Printf("Adaptive sampling completed: used %d samples out of %d hands (%.1f%%), average equity: %.2f%%, total equities calculated: %d",
			samplesUsed, len(formattedOpponentHands), 
			float64(samplesUsed)/float64(len(formattedOpponentHands))*100,
			avgEquity, len(equities))
		
		return equities, samplesUsed, nil
	} else if config.UseMonteCarloEquity {
		// Monte Carlo法を使用（各ハンドに対して個別に計算）
		log.Printf("Using Monte Carlo equity calculation (mode: %s)", config.MonteCarloMode)
		
		// Adaptive設定を作成
		var adaptiveConfig pkrlib.EquityCalculationConfig
		switch config.MonteCarloMode {
		case "FAST":
			adaptiveConfig = pkrlib.EquityCalculationConfig{
				MaxIterations:    pkrlib.FAST_ITERATIONS,
				TargetPrecision:  3.0, // ±3%誤差
				MinIterations:    500,
				ConvergenceCheck: 100,
			}
		case "ACCURATE":
			adaptiveConfig = pkrlib.EquityCalculationConfig{
				MaxIterations:    pkrlib.ACCURATE_ITERATIONS,
				TargetPrecision:  0.5, // ±0.5%誤差
				MinIterations:    2000,
				ConvergenceCheck: 200,
			}
		case "NORMAL":
			fallthrough
		default:
			adaptiveConfig = pkrlib.GetDefaultAdaptiveConfig()
		}
		
		// 各相手ハンドに対してAdaptive計算を実行
		equities := make(map[string]float64)
		totalIterations := 0
		
		for _, opponentHand := range formattedOpponentHands {
			if pkrlib.HasCardDuplicates(yourHand, opponentHand, flop) {
				continue
			}
			
			villainHandStr := ""
			for _, card := range opponentHand {
				villainHandStr += card.String()
			}
			
			equity, iterations, err := pkrlib.CalculateHandVsHandEquityAdaptive(yourHand, opponentHand, flop, adaptiveConfig)
			if err == nil && equity != -1 {
				equities[villainHandStr] = equity
				totalIterations += iterations
			}
		}
		
		log.Printf("Monte Carlo calculation completed with average %d iterations per hand", totalIterations/len(equities))
		return equities, 0, nil
	} else {
		// 従来のExhaustive法を使用
		if config.EnableParallelProcessing {
			// 並列処理が有効な場合は並列計算関数を使用
			log.Printf("Using exhaustive equity calculation with parallel processing")
			equities, err := pkrlib.CalculateHandVsRangeEquityParallel(yourHand, formattedOpponentHands, flop)
			return equities, 0, err
		} else {
			// 並列処理が無効な場合は非並列計算関数を使用
			// 注: pkrlib.CalculateHandVsRangeEquityという非並列版の関数が存在しない場合は、
			// 並列版の関数を使用します
			log.Printf("Using exhaustive equity calculation")
			equities, err := pkrlib.CalculateHandVsRangeEquityParallel(yourHand, formattedOpponentHands, flop)
			return equities, 0, err
		}
	}
}
