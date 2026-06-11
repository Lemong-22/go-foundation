-- Practical caveat: after authors embed practice blocks, this rollback can only
-- run after those blocks are removed because Phase C practice references are data.
DROP INDEX IF EXISTS content_blocks_practice_ref_idx;

ALTER TABLE content_blocks
    DROP CONSTRAINT IF EXISTS content_blocks_practice_ref_fkey;

ALTER TABLE content_blocks
    DROP COLUMN IF EXISTS practice_ref;

ALTER TABLE content_blocks
    DROP CONSTRAINT IF EXISTS content_blocks_kind_check;

ALTER TABLE content_blocks
    ADD CONSTRAINT content_blocks_kind_check
    CHECK (kind IN ('text', 'video', 'quiz'));

DROP TABLE IF EXISTS practice_test_cases;

DROP TABLE IF EXISTS practices;
