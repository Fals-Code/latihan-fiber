INSERT INTO permissions (name, description)
VALUES
    ('student:list', 'Melihat daftar seluruh mahasiswa'),
    ('student:read:any', 'Melihat data mahasiswa lain'),
    ('student:create', 'Membuat data mahasiswa'),
    ('student:update:any', 'Mengubah data mahasiswa lain'),
    ('student:delete', 'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name)
VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT (role_name, permission_name) DO NOTHING;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'students'
          AND column_name = 'owner_id'
    ) THEN
        ALTER TABLE students ADD COLUMN owner_id INTEGER;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'students'::regclass
          AND conname = 'students_owner_id_fkey'
    ) THEN
        ALTER TABLE students
            ADD CONSTRAINT students_owner_id_fkey
            FOREIGN KEY (owner_id) REFERENCES users(id);
    END IF;
END
$$;
