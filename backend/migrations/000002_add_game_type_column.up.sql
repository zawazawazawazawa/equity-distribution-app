-- game_typeカラムを追加
ALTER TABLE daily_quiz_results ADD COLUMN game_type VARCHAR(20) NOT NULL DEFAULT '4card_plo';

-- game_typeカラム用のインデックスを作成
CREATE INDEX IF NOT EXISTS idx_daily_quiz_results_game_type ON daily_quiz_results(game_type);
