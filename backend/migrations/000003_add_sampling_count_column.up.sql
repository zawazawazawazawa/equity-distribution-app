-- Add sampling count column to daily_quiz_results table
ALTER TABLE daily_quiz_results 
ADD COLUMN sampling_count INTEGER DEFAULT NULL;

-- Add comment for the column
COMMENT ON COLUMN daily_quiz_results.sampling_count IS 'Number of samples used in adaptive sampling calculation';

-- Create index on sampling_count for potential filtering/sorting
CREATE INDEX idx_daily_quiz_results_sampling_count ON daily_quiz_results(sampling_count);