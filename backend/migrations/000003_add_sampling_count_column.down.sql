-- Drop the index
DROP INDEX IF EXISTS idx_daily_quiz_results_sampling_count;

-- Remove the sampling_count column
ALTER TABLE daily_quiz_results
DROP COLUMN IF EXISTS sampling_count;