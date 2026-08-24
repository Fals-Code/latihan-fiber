CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim BIGINT NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL CHECK (grade >= 0 AND grade <= 100),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));