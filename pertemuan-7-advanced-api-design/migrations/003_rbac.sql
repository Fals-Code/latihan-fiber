CREATE TABLE IF NOT EXISTS roles (
    name VARCHAR(20) PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO roles (name, description)
VALUES
    ('admin', 'Akses penuh terhadap seluruh data dan pengaturan'),
    ('staff', 'Boleh melihat data seluruh user, tetapi tidak boleh mengubah'),
    ('user', 'Hanya boleh mengelola datanya sendiri')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, description)
VALUES
    ('user:list', 'Melihat daftar seluruh user'),
    ('user:read:any', 'Melihat data user lain'),
    ('user:update:any', 'Mengubah data user lain'),
    ('user:delete', 'Menghapus user'),
    ('role:assign', 'Mengubah role user')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name)
VALUES
    ('admin', 'user:list'),
    ('admin', 'user:read:any'),
    ('admin', 'user:update:any'),
    ('admin', 'user:delete'),
    ('admin', 'role:assign'),
    ('staff', 'user:list'),
    ('staff', 'user:read:any')
ON CONFLICT (role_name, permission_name) DO NOTHING;

UPDATE users
SET role = 'user'
WHERE role NOT IN (SELECT name FROM roles);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'users'::regclass
          AND conname = 'users_role_fkey'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT users_role_fkey
            FOREIGN KEY (role)
            REFERENCES roles(name)
            ON UPDATE CASCADE;
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);
