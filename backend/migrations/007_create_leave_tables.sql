-- Modul 5: Smart Leave Management
-- Migration: 007_create_leave_tables.sql

-- Tabel jenis cuti (tahunan, sakit, izin, melahirkan, dll)
CREATE TABLE IF NOT EXISTS leave_types (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    max_days    INT NOT NULL DEFAULT 0,
    is_paid     BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabel saldo cuti per karyawan per tahun
CREATE TABLE IF NOT EXISTS leave_balances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leave_type_id   UUID NOT NULL REFERENCES leave_types(id) ON DELETE CASCADE,
    year            INT NOT NULL,
    total_days      INT NOT NULL DEFAULT 0,
    used_days       INT NOT NULL DEFAULT 0,
    remaining_days  INT GENERATED ALWAYS AS (total_days - used_days) STORED,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, leave_type_id, year)
);

-- Tabel pengajuan cuti dengan state machine
CREATE TABLE IF NOT EXISTS leave_requests (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leave_type_id     UUID NOT NULL REFERENCES leave_types(id),
    start_date        DATE NOT NULL,
    end_date          DATE NOT NULL,
    total_days        INT NOT NULL,
    reason            TEXT NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    approved_by       UUID REFERENCES users(id),
    approved_at       TIMESTAMPTZ,
    rejection_reason  TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk performa query
CREATE INDEX IF NOT EXISTS idx_leave_requests_user_id ON leave_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_leave_requests_status ON leave_requests(status);
CREATE INDEX IF NOT EXISTS idx_leave_requests_start_date ON leave_requests(start_date);
CREATE INDEX IF NOT EXISTS idx_leave_balances_user_id ON leave_balances(user_id);
CREATE INDEX IF NOT EXISTS idx_leave_balances_year ON leave_balances(year);

-- Seed data: jenis cuti default
INSERT INTO leave_types (name, max_days, is_paid, description) VALUES
    ('Cuti Tahunan', 12, true, 'Cuti tahunan yang diberikan kepada karyawan tetap'),
    ('Cuti Sakit', 14, true, 'Cuti karena sakit dengan surat dokter'),
    ('Cuti Izin', 3, true, 'Cuti izin keperluan mendesak'),
    ('Cuti Melahirkan', 90, true, 'Cuti melahirkan untuk karyawan perempuan'),
    ('Cuti Tidak Berbayar', 30, false, 'Cuti tanpa gaji atas persetujuan manajemen')
ON CONFLICT (name) DO NOTHING;
