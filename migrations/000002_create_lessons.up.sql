CREATE TABLE IF NOT EXISTS lessons (
    id VARCHAR(50) PRIMARY KEY,
    course_id VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    content TEXT,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Ini si Kunci Asing (Foreign Key) yang mengikat tabel lessons ke tabel courses
    CONSTRAINT fk_lessons_course FOREIGN KEY (course_id) 
        REFERENCES courses(id) ON DELETE CASCADE
);