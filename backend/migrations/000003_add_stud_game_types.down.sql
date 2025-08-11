-- Remove stud game type support
-- This migration reverts the changes made in the up migration

-- Remove the check constraint if it exists
ALTER TABLE daily_quiz_results DROP CONSTRAINT IF EXISTS chk_game_type;

-- Remove the comment
COMMENT ON COLUMN daily_quiz_results.game_type IS NULL;