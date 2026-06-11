CREATE TABLE tests (
    id UUID PRIMARY KEY,
    course_id UUID NOT NULL,
    title TEXT NOT NULL,
    time_limit_minutes INTEGER,
    pass_threshold DOUBLE PRECISION NOT NULL DEFAULT 0.7,
    solution_zip_provider TEXT,
    solution_zip_locator TEXT,
    solution_video_provider TEXT,
    solution_video_locator TEXT,
    solution_video_caption TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tests_course_id_fkey
        FOREIGN KEY (course_id)
        REFERENCES courses(id)
        ON DELETE CASCADE,
    CONSTRAINT tests_time_limit_minutes_check
        CHECK (time_limit_minutes IS NULL OR time_limit_minutes > 0),
    CONSTRAINT tests_pass_threshold_check
        CHECK (pass_threshold >= 0 AND pass_threshold <= 1),
    CONSTRAINT tests_solution_group_check
        CHECK (
            (
                solution_zip_provider IS NULL
                AND solution_zip_locator IS NULL
                AND solution_video_provider IS NULL
                AND solution_video_locator IS NULL
                AND solution_video_caption IS NULL
            )
            OR
            (
                solution_zip_provider IS NOT NULL
                AND solution_zip_locator IS NOT NULL
                AND solution_video_provider IS NOT NULL
                AND solution_video_locator IS NOT NULL
            )
        )
);

CREATE TABLE test_items (
    id UUID PRIMARY KEY,
    test_id UUID NOT NULL,
    kind TEXT NOT NULL,
    position INTEGER NOT NULL,
    choice_type TEXT,
    choice_prompt TEXT,
    choice_options JSONB,
    choice_correct_indices JSONB,
    choice_explanation TEXT,
    coding_language TEXT,
    coding_prompt TEXT,
    starter_code TEXT,
    coding_solution TEXT,
    coding_test_cases JSONB,
    CONSTRAINT test_items_test_id_fkey
        FOREIGN KEY (test_id)
        REFERENCES tests(id)
        ON DELETE CASCADE,
    CONSTRAINT test_items_kind_check
        CHECK (kind IN ('choice', 'coding')),
    CONSTRAINT test_items_position_check
        CHECK (position >= 0),
    CONSTRAINT test_items_choice_type_check
        CHECK (choice_type IS NULL OR choice_type IN ('single', 'multiple')),
    CONSTRAINT test_items_coding_language_check
        CHECK (coding_language IS NULL OR coding_language IN ('javascript', 'typescript', 'golang', 'rust')),
    CONSTRAINT test_items_body_shape_check
        CHECK (
            (
                kind = 'choice'
                AND choice_type IS NOT NULL
                AND choice_prompt IS NOT NULL
                AND choice_options IS NOT NULL
                AND choice_correct_indices IS NOT NULL
                AND choice_explanation IS NOT NULL
                AND coding_language IS NULL
                AND coding_prompt IS NULL
                AND starter_code IS NULL
                AND coding_solution IS NULL
                AND coding_test_cases IS NULL
            )
            OR
            (
                kind = 'coding'
                AND choice_type IS NULL
                AND choice_prompt IS NULL
                AND choice_options IS NULL
                AND choice_correct_indices IS NULL
                AND choice_explanation IS NULL
                AND coding_language IS NOT NULL
                AND coding_prompt IS NOT NULL
                AND starter_code IS NOT NULL
                AND coding_solution IS NOT NULL
                AND coding_test_cases IS NOT NULL
            )
        ),
    CONSTRAINT test_items_position_unique
        UNIQUE (test_id, position)
);

CREATE INDEX test_items_test_position_idx
    ON test_items (test_id, position);
