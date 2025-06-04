package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// PostgresConfig はPostgreSQL接続設定を表します
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// GetPostgresConnection はPostgreSQLへの接続を確立します
// 環境変数から接続情報を取得することもできます
func GetPostgresConnection(config PostgresConfig) (*sql.DB, error) {
	// 環境変数から値を取得（設定されている場合）
	host := getEnvOrDefault("POSTGRES_HOST", config.Host)
	port := getEnvIntOrDefault("POSTGRES_PORT", config.Port)
	user := getEnvOrDefault("POSTGRES_USER", config.User)
	password := getEnvOrDefault("POSTGRES_PASSWORD", config.Password)
	dbName := getEnvOrDefault("POSTGRES_DBNAME", config.DBName)

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	// 接続テスト
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	return db, nil
}

// InsertDailyQuizResult は計算結果をPostgreSQLに保存します
func InsertDailyQuizResult(db *sql.DB, date time.Time, scenario string, heroHand string, flop string, result string, averageEquity float64, gameType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO daily_quiz_results (date, scenario, hero_hand, flop, result, average_equity, game_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int
	err := db.QueryRowContext(ctx, query, date, scenario, heroHand, flop, result, averageEquity, gameType).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert data into PostgreSQL: %v", err)
	}

	log.Printf("Successfully inserted data into PostgreSQL with ID: %d", id)
	return nil
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

// GetDailyQuizResultsByDate は指定された日付のクイズ結果を取得します
func GetDailyQuizResultsByDate(db *sql.DB, date time.Time) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, date, scenario, hero_hand, flop, result, average_equity, created_at
		FROM daily_quiz_results
		WHERE date = $1
		ORDER BY id
	`

	rows, err := db.QueryContext(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query data from PostgreSQL: %v", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var date time.Time
		var scenario, heroHand, flop, result string
		var averageEquity float64
		var createdAt time.Time

		if err := rows.Scan(&id, &date, &scenario, &heroHand, &flop, &result, &averageEquity, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		// JSONデータをパース
		var resultData interface{}
		if result != "" {
			if err := json.Unmarshal([]byte(result), &resultData); err != nil {
				log.Printf("Warning: Failed to parse JSON result: %v", err)
				// エラーがあってもデータは返す
			}
		}

		// 結果をマップに格納
		item := map[string]interface{}{
			"id":             id,
			"date":           date.Format("2006-01-02"),
			"scenario":       scenario,
			"hero_hand":      heroHand,
			"flop":           flop,
			"result":         resultData,
			"average_equity": averageEquity,
			"created_at":     createdAt,
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}
