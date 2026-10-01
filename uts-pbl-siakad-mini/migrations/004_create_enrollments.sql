CREATE TABLE IF NOT EXISTS enrollments (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE RESTRICT,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    tahun_akademik VARCHAR(20) NOT NULL CHECK (tahun_akademik ~ '^[0-9]{4}/[0-9]{4}-(Ganjil|Genap)$'),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT enrollments_student_course_year_unique UNIQUE (student_id, course_id, tahun_akademik)
);

CREATE INDEX IF NOT EXISTS idx_enrollments_course_id ON enrollments(course_id);
