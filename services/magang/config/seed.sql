-- Setup database untuk magang-be
-- Jalankan di psql atau pgAdmin

-- Kalau belum ada databasenya, buat dulu:
-- CREATE DATABASE magang_dev;

CREATE SCHEMA IF NOT EXISTS magang;

CREATE TABLE IF NOT EXISTS magang.users (
    id BIGSERIAL PRIMARY KEY,
    no_emp VARCHAR(50),
    id_emp VARCHAR(50),
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    title VARCHAR(150),
    company VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    section VARCHAR(100),
    department VARCHAR(100),
    division VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Migrasi jika tabel sudah ada (tambah kolom baru)
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS no_emp VARCHAR(50);
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS id_emp VARCHAR(50);
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS title VARCHAR(150);
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS company VARCHAR(255);
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS section VARCHAR(100);
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS department VARCHAR(100);
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS division VARCHAR(100);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_role_check'
    ) THEN
        ALTER TABLE magang.users
        ADD CONSTRAINT users_role_check
        CHECK (role IN ('admin', 'user', 'approval', 'external_audit'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS magang.items (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed user (password: admin123)
INSERT INTO magang.users (username, password, full_name, role)
VALUES ('admin', 'admin123', 'Admin Magang', 'admin')
ON CONFLICT (username) DO NOTHING;

-- Seed normal user (password: user123)
INSERT INTO magang.users (username, password, full_name, role)
VALUES ('user', 'user123', 'User Magang', 'user')
ON CONFLICT (username) DO NOTHING;
