package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestInsertDailyQuizResult(t *testing.T) {
	// SQLモックを作成
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// テストデータ
	testDate := time.Date(2024, 6, 5, 0, 0, 0, 0, time.UTC)
	scenario := "SRP UTG vs BB"
	heroHand := "AhAsKdQc"
	flop := "2d3cJc"
	result := `[{"villain_hand":"KsKcQdJh","equity":0.35}]`
	averageEquity := 65.50
	gameType := "4card_plo"

	t.Run("成功ケース", func(t *testing.T) {
		// 期待されるクエリとレスポンスを設定
		mock.ExpectQuery(`INSERT INTO daily_quiz_results \(date, scenario, hero_hand, flop, result, average_equity, game_type\)`).
			WithArgs(testDate, scenario, heroHand, flop, result, averageEquity, gameType).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// 関数を実行
		err := InsertDailyQuizResult(db, testDate, scenario, heroHand, flop, result, averageEquity, gameType)

		// アサーション
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("データベースエラー", func(t *testing.T) {
		// エラーを返すクエリを設定
		mock.ExpectQuery(`INSERT INTO daily_quiz_results \(date, scenario, hero_hand, flop, result, average_equity, game_type\)`).
			WithArgs(testDate, scenario, heroHand, flop, result, averageEquity, gameType).
			WillReturnError(sql.ErrConnDone)

		// 関数を実行
		err := InsertDailyQuizResult(db, testDate, scenario, heroHand, flop, result, averageEquity, gameType)

		// アサーション
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert data into PostgreSQL")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("game_typeパラメータが正しく渡される", func(t *testing.T) {
		// 異なるgame_typeでテスト
		differentGameType := "holdem"
		
		mock.ExpectQuery(`INSERT INTO daily_quiz_results \(date, scenario, hero_hand, flop, result, average_equity, game_type\)`).
			WithArgs(testDate, scenario, heroHand, flop, result, averageEquity, differentGameType).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

		// 関数を実行
		err := InsertDailyQuizResult(db, testDate, scenario, heroHand, flop, result, averageEquity, differentGameType)

		// アサーション
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetDailyQuizResultsByDate(t *testing.T) {
	// SQLモックを作成
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// テストデータ
	testDate := time.Date(2024, 6, 5, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, 6, 5, 12, 0, 0, 0, time.UTC)

	t.Run("成功ケース", func(t *testing.T) {
		// モックのレスポンスを設定
		rows := sqlmock.NewRows([]string{"id", "date", "scenario", "hero_hand", "flop", "result", "average_equity", "created_at"}).
			AddRow(1, testDate, "SRP UTG vs BB", "AhAsKdQc", "2d3cJc", `[{"villain_hand":"KsKcQdJh","equity":0.35}]`, 65.50, createdAt).
			AddRow(2, testDate, "SRP BTN vs BB", "KhKsQdJc", "5h6s7c", `[{"villain_hand":"AhAsKdQc","equity":0.65}]`, 45.20, createdAt)

		mock.ExpectQuery(`SELECT id, date, scenario, hero_hand, flop, result, average_equity, created_at FROM daily_quiz_results WHERE date = \$1 ORDER BY id`).
			WithArgs(testDate).
			WillReturnRows(rows)

		// 関数を実行
		results, err := GetDailyQuizResultsByDate(db, testDate)

		// アサーション
		assert.NoError(t, err)
		assert.Len(t, results, 2)

		// 最初のレコードをチェック
		assert.Equal(t, 1, results[0]["id"])
		assert.Equal(t, "2024-06-05", results[0]["date"])
		assert.Equal(t, "SRP UTG vs BB", results[0]["scenario"])
		assert.Equal(t, "AhAsKdQc", results[0]["hero_hand"])
		assert.Equal(t, "2d3cJc", results[0]["flop"])
		assert.Equal(t, 65.50, results[0]["average_equity"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("結果が見つからない場合", func(t *testing.T) {
		// 空の結果を返すモックを設定
		rows := sqlmock.NewRows([]string{"id", "date", "scenario", "hero_hand", "flop", "result", "average_equity", "created_at"})

		mock.ExpectQuery(`SELECT id, date, scenario, hero_hand, flop, result, average_equity, created_at FROM daily_quiz_results WHERE date = \$1 ORDER BY id`).
			WithArgs(testDate).
			WillReturnRows(rows)

		// 関数を実行
		results, err := GetDailyQuizResultsByDate(db, testDate)

		// アサーション
		assert.NoError(t, err)
		assert.Len(t, results, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("データベースエラー", func(t *testing.T) {
		// エラーを返すクエリを設定
		mock.ExpectQuery(`SELECT id, date, scenario, hero_hand, flop, result, average_equity, created_at FROM daily_quiz_results WHERE date = \$1 ORDER BY id`).
			WithArgs(testDate).
			WillReturnError(sql.ErrConnDone)

		// 関数を実行
		results, err := GetDailyQuizResultsByDate(db, testDate)

		// アサーション
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "failed to query data from PostgreSQL")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}