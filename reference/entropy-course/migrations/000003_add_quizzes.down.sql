-- Practical caveat: after authors embed quiz blocks, this rollback can only
-- run after those blocks are removed because Phase B quiz references are data.
DROP INDEX IF EXISTS content_blocks_quiz_ref_idx;

ALTER TABLE content_blocks
    DROP CONSTRAINT IF EXISTS content_blocks_quiz_ref_fkey;

ALTER TABLE content_blocks
    DROP COLUMN IF EXISTS quiz_ref;

ALTER TABLE content_blocks
    DROP CONSTRAINT IF EXISTS content_blocks_kind_check;

ALTER TABLE content_blocks
    ADD CONSTRAINT content_blocks_kind_check
    CHECK (kind IN ('text', 'video'));

DROP TABLE IF EXISTS quiz_questions;

DROP TABLE IF EXISTS quizzes;
