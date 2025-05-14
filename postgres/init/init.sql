-- バッチ処理の結果を保存するテーブル
CREATE TABLE IF NOT EXISTS daily_quiz_results (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    scenario VARCHAR(255) NOT NULL,
    hero_hand VARCHAR(255) NOT NULL,
    flop VARCHAR(255) NOT NULL,
    result TEXT,
    average_equity DECIMAL(5,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- インデックスの作成
CREATE INDEX IF NOT EXISTS idx_daily_quiz_results_date ON daily_quiz_results(date);
CREATE INDEX IF NOT EXISTS idx_daily_quiz_results_scenario ON daily_quiz_results(scenario);
CREATE INDEX IF NOT EXISTS idx_daily_quiz_results_hero_hand ON daily_quiz_results(hero_hand);
CREATE INDEX IF NOT EXISTS idx_daily_quiz_results_flop ON daily_quiz_results(flop);