-- インデックスの削除
DROP INDEX IF EXISTS idx_daily_quiz_results_flop;
DROP INDEX IF EXISTS idx_daily_quiz_results_hero_hand;
DROP INDEX IF EXISTS idx_daily_quiz_results_scenario;
DROP INDEX IF EXISTS idx_daily_quiz_results_date;

-- テーブルの削除
DROP TABLE IF EXISTS daily_quiz_results;
