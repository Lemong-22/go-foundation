CREATE TABLE content_blocks (
    id UUID PRIMARY KEY,
    lesson_id UUID NOT NULL,
    kind TEXT NOT NULL,
    position INTEGER NOT NULL,
    text_markdown TEXT,
    video_provider TEXT,
    video_locator TEXT,
    video_caption TEXT,
    CONSTRAINT content_blocks_lesson_id_fkey
        FOREIGN KEY (lesson_id)
        REFERENCES lessons(id)
        ON DELETE CASCADE,
    CONSTRAINT content_blocks_kind_check CHECK (kind IN ('text', 'video')),
    CONSTRAINT content_blocks_position_check CHECK (position >= 0),
    CONSTRAINT content_blocks_lesson_position_key UNIQUE (lesson_id, position)
);

CREATE INDEX content_blocks_lesson_position_idx ON content_blocks (lesson_id, position);

INSERT INTO content_blocks (
    id,
    lesson_id,
    kind,
    position,
    text_markdown
)
SELECT
    id,
    id,
    'text',
    0,
    content
FROM lessons;
