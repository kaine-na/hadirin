-- 001_create_users.sql
-- Tabel utama untuk semua user (super_admin, hr_admin, manager, karyawan)

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL DEFAULT gen_random_uuid(), -- multi-tenant, default untuk single-tenant
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role        VARCHAR(50) NOT NULL DEFAULT 'karyawan',
                -- 'super_admin' | 'hr_admin' | 'manager' | 'karyawan'
    department  VARCHAR(100),
    position    VARCHAR(100),
    nik         VARCHAR(50),
    photo_url   VARCHAR(500),
    joined_at   DATE,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
