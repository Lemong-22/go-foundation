CREATE TABLE practices (
    id UUID PRIMARY KEY,
    course_id UUID NOT NULL,
    title TEXT NOT NULL,
    language TEXT NOT NULL,
    prompt TEXT NOT NULL,
    starter_code TEXT NOT NULL DEFAULT '',
    solution TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT practices_course_id_fkey
        FOREIGN KEY (course_id)
        REFERENCES courses(id)
        ON DELETE CASCADE,
    CONSTRAINT practices_language_check
        CHECK (language IN ('javascript', 'typescript', 'golang', 'rust'))
);

CREATE TABLE practice_test_cases (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL,
    stdin TEXT NOT NULL DEFAULT '',
    expected_stdout TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL,
    CONSTRAINT practice_test_cases_practice_id_fkey
        FOREIGN KEY (practice_id)
        REFERENCES practices(id)
        ON DELETE CASCADE,
    CONSTRAINT practice_test_cases_position_check
        CHECK (position >= 0),
    CONSTRAINT practice_test_cases_position_unique
        UNIQUE (practice_id, position)
);

CREATE INDEX practice_test_cases_practice_position_idx
    ON practice_test_cases (practice_id, position);

ALTER TABLE content_blocks
    DROP CONSTRAINT content_blocks_kind_check;

ALTER TABLE content_blocks
    ADD CONSTRAINT content_blocks_kind_check
    CHECK (kind IN ('text', 'video', 'quiz', 'practice'));

ALTER TABLE content_blocks
    ADD COLUMN practice_ref UUID;

ALTER TABLE content_blocks
    ADD CONSTRAINT content_blocks_practice_ref_fkey
    FOREIGN KEY (practice_ref)
    REFERENCES practices(id)
    ON DELETE RESTRICT;

CREATE INDEX content_blocks_practice_ref_idx
    ON content_blocks (practice_ref);
