CREATE TABLE quizzes (
    id UUID PRIMARY KEY,
    course_id UUID NOT NULL,
    title TEXT NOT NULL,
    pass_threshold DOUBLE PRECISION NOT NULL DEFAULT 0.7,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT quizzes_course_id_fkey
        FOREIGN KEY (course_id)
        REFERENCES courses(id)
        ON DELETE CASCADE,
    CONSTRAINT quizzes_pass_threshold_check
        CHECK (pass_threshold >= 0 AND pass_threshold <= 1)
);

CREATE TABLE quiz_questions (
    id UUID PRIMARY KEY,
    quiz_id UUID NOT NULL,
    type TEXT NOT NULL,
    prompt TEXT NOT NULL,
    options JSONB NOT NULL,
    correct_indices JSONB NOT NULL,
    explanation TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL,
    CONSTRAINT quiz_questions_quiz_id_fkey
        FOREIGN KEY (quiz_id)
        REFERENCES quizzes(id)
        ON DELETE CASCADE,
    CONSTRAINT quiz_questions_type_check
        CHECK (type IN ('single', 'multiple')),
    CONSTRAINT quiz_questions_position_check
        CHECK (position >= 0),
    CONSTRAINT quiz_questions_position_unique
        UNIQUE (quiz_id, position)
);

CREATE INDEX quiz_questions_quiz_position_idx
    ON quiz_questions (quiz_id, position);

ALTER TABLE content_blocks
    DROP CONSTRAINT content_blocks_kind_check;

ALTER TABLE content_blocks
    ADD CONSTRAINT content_blocks_kind_check
    CHECK (kind IN ('text', 'video', 'quiz'));

ALTER TABLE content_blocks
    ADD COLUMN quiz_ref UUID;

ALTER TABLE content_blocks
    ADD CONSTRAINT content_blocks_quiz_ref_fkey
    FOREIGN KEY (quiz_ref)
    REFERENCES quizzes(id)
    ON DELETE RESTRICT;

CREATE INDEX content_blocks_quiz_ref_idx
    ON content_blocks (quiz_ref);
