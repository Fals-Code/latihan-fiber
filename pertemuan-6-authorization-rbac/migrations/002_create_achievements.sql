CREATE TABLE IF NOT EXISTS achievements (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    rank int NOT NULL CHECK (rank > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS achievement_name_lower_idx
    ON achievements (LOWER(name))
    ;