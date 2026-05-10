-- 002_create_attendances.sql
-- Tabel absensi karyawan dengan unique constraint per user per hari

CREATE TABLE IF NOT EXISTS attendances (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date        DATE NOT NULL,
    clock_in    TIMESTAMPTZ,
    clock_out   TIMESTAMPTZ,
    status      VARCHAR(20) NOT NULL DEFAULT 'hadir',
                -- 'hadir' | 'terlambat' | 'izin' | 'sakit' | 'alpha'
    notes       TEXT,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    created_by  UUID REFERENCES users(id),  -- jika diinput HR
    updated_by  UUID REFERENCES users(id),  -- jika di-override HR
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_attendances_user_id ON attendances(user_id);
CREATE INDEX IF NOT EXISTS idx_attendances_date ON attendances(date);
