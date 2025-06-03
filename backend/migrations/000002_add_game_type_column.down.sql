-- game_typeカラム用のインデックスを削除
DROP INDEX IF EXISTS idx_daily_quiz_results_game_type;

-- game_typeカラムを削除
ALTER TABLE daily_quiz_results DROP COLUMN game_type;
