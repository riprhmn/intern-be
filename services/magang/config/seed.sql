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
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS code_name VARCHAR(50);
ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS level_rank INT NOT NULL DEFAULT 4; -- 1: DivHead, 2: DeptHead, 3: SecHead, 4: Staff
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

-- Tabel Master Mapping Approval Per Seksi & Tipe Dokumen
CREATE TABLE IF NOT EXISTS magang.approval_mappings (
    id BIGSERIAL PRIMARY KEY,
    approval_type VARCHAR(10) NOT NULL, -- 'CN' atau 'ANCR'
    section_code VARCHAR(100) NOT NULL,
    seq_1 VARCHAR(50), -- CODE_NAME Stage 1
    seq_2 VARCHAR(50), -- CODE_NAME Stage 2
    seq_3 VARCHAR(50), -- CODE_NAME Stage 3
    seq_4 VARCHAR(50), -- CODE_NAME Stage 4
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tabel Delegasi Cuti
CREATE TABLE IF NOT EXISTS magang.approval_delegations (
    id BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT NOT NULL REFERENCES magang.users(id),
    to_user_id BIGINT NOT NULL REFERENCES magang.users(id),
    approval_type VARCHAR(10) NOT NULL DEFAULT 'ALL',
    section_code VARCHAR(100) NOT NULL DEFAULT 'ALL',
    starts_at TIMESTAMP,
    ends_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tabel Relasional Transaksi Approval (Header & Detail)
CREATE TABLE IF NOT EXISTS magang.transaction_headers (
    id BIGSERIAL PRIMARY KEY,
    doc_number VARCHAR(100) NOT NULL UNIQUE,
    approval_type VARCHAR(10) NOT NULL, -- 'CN' atau 'ANCR'
    section_code VARCHAR(100) NOT NULL,
    current_seq INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'IN PROGRESS', -- 'IN PROGRESS', 'APPROVED', 'REJECTED'
    created_by BIGINT NOT NULL REFERENCES magang.users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS magang.transaction_details (
    id BIGSERIAL PRIMARY KEY,
    header_id BIGINT NOT NULL REFERENCES magang.transaction_headers(id) ON DELETE CASCADE,
    seq_stage INT NOT NULL,
    target_code_name VARCHAR(50) NOT NULL,
    executed_by_user_id BIGINT REFERENCES magang.users(id),
    delegated_from_user_id BIGINT REFERENCES magang.users(id),
    action VARCHAR(20), -- 'APPROVE', 'REJECT'
    notes TEXT,
    executed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed user (password: admin123)
INSERT INTO magang.users (username, password, full_name, role, code_name, level_rank)
VALUES ('admin', 'admin123', 'Admin Magang', 'admin', 'AD', 1)
ON CONFLICT (username) DO NOTHING;

-- Seed normal user (password: user123)
INSERT INTO magang.users (username, password, full_name, role, code_name, level_rank)
VALUES ('user', 'user123', 'User Magang', 'user', 'AG', 4)
ON CONFLICT (username) DO NOTHING;

