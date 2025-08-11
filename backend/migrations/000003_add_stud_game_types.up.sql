-- Add support for stud game types in game_type column
-- This migration updates the game_type column to support new stud game variants

-- First, we need to check if we need to modify the column constraints
-- Since game_type is VARCHAR(20), it should already support the new values:
-- - 'razz' (4 chars)
-- - '7card_stud_high' (15 chars)
-- - '7card_stud_highlow8' (19 chars)

-- Add a check constraint to validate game types (if not already present)
-- This ensures only valid game types can be inserted
DO $$
BEGIN
    -- Check if constraint already exists
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.constraint_column_usage 
        WHERE constraint_name = 'chk_game_type'
    ) THEN
        ALTER TABLE daily_quiz_results 
        ADD CONSTRAINT chk_game_type 
        CHECK (game_type IN (
            '4card_plo',
            '5card_plo',
            'holdem',
            'razz',
            '7card_stud_high',
            '7card_stud_highlow8'
        ));
    END IF;
END $$;

-- Add comment to document the new game types
COMMENT ON COLUMN daily_quiz_results.game_type IS 'Game type: 4card_plo, 5card_plo, holdem, razz, 7card_stud_high, 7card_stud_highlow8';