package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
func GetPostgresConnection(config PostgresConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

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
func InsertDailyQuizResult(db *sql.DB, date time.Time, scenario string, heroHand string, flop string, result string, averageEquity float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO daily_quiz_results (date, scenario, hero_hand, flop, result, average_equity)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := db.QueryRowContext(ctx, query, date, scenario, heroHand, flop, result, averageEquity).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert data into PostgreSQL: %v", err)
	}

	log.Printf("Successfully inserted data into PostgreSQL with ID: %d", id)
	return nil
}
