CREATE TABLE courses (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    instructor_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT courses_slug_key UNIQUE (slug),
    CONSTRAINT courses_status_check CHECK (status IN ('draft', 'published'))
);

CREATE TABLE lessons (
    id UUID PRIMARY KEY,
    course_id UUID NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    "order" INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lessons_course_id_fkey
        FOREIGN KEY (course_id)
        REFERENCES courses(id)
        ON DELETE CASCADE,
    CONSTRAINT lessons_order_check CHECK ("order" >= 0)
);

CREATE INDEX lessons_course_order_idx ON lessons (course_id, "order");
