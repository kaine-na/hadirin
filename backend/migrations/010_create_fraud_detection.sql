-- Migration 010: Fraud Detection Tables
-- attendance_photos: foto selfie saat clock-in
-- fraud_logs: log deteksi fraud dengan bukti
-- device_fingerprints: fingerprint device per user

CREATE TABLE IF NOT EXISTS attendance_photos (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id UUID NOT NULL REFERENCES attendances(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_path   TEXT NOT NULL,
    file_size   INTEGER NOT NULL,
    mime_type   VARCHAR(50) NOT NULL DEFAULT 'image/jpeg',
    -- Hasil liveness check dari AI
    is_live_face    BOOLEAN,
    liveness_score  DECIMAL(5,4),
    liveness_notes  TEXT,
    -- GPS saat foto diambil
    latitude    DECIMAL(10,8),
    longitude   DECIMAL(11,8),
    gps_accuracy DECIMAL(8,2),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS device_fingerprints (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_hash VARCHAR(64) NOT NULL,
    user_agent  TEXT,
    platform    VARCHAR(100),
    first_seen  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_trusted  BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE(user_id, device_hash)
);

CREATE TABLE IF NOT EXISTS fraud_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id   UUID REFERENCES attendances(id) ON DELETE SET NULL,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fraud_type      VARCHAR(50) NOT NULL,
    -- Tipe: gps_accuracy, mock_location, velocity_check, anomaly_time,
    --        anomaly_location, anomaly_device, liveness_fail
    severity        VARCHAR(20) NOT NULL DEFAULT 'medium',
    -- Severity: low, medium, high, critical
    description     TEXT NOT NULL,
    evidence        JSONB,
    -- Bukti: koordinat, foto path, device info, dll
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- Status: pending, dismissed, confirmed
    reviewed_by     UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at     TIMESTAMPTZ,
    review_notes    TEXT,
    -- AI analysis
    ai_analysis     TEXT,
    ai_confidence   DECIMAL(5,4),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk query performa
CREATE INDEX IF NOT EXISTS idx_attendance_photos_attendance_id ON attendance_photos(attendance_id);
CREATE INDEX IF NOT EXISTS idx_attendance_photos_user_id ON attendance_photos(user_id);
CREATE INDEX IF NOT EXISTS idx_device_fingerprints_user_id ON device_fingerprints(user_id);
CREATE INDEX IF NOT EXISTS idx_fraud_logs_user_id ON fraud_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_fraud_logs_status ON fraud_logs(status);
CREATE INDEX IF NOT EXISTS idx_fraud_logs_created_at ON fraud_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_fraud_logs_fraud_type ON fraud_logs(fraud_type);

-- Tambahkan kolom GPS ke tabel attendances jika belum ada
ALTER TABLE attendances
    ADD COLUMN IF NOT EXISTS latitude    DECIMAL(10,8),
    ADD COLUMN IF NOT EXISTS longitude   DECIMAL(11,8),
    ADD COLUMN IF NOT EXISTS gps_accuracy DECIMAL(8,2),
    ADD COLUMN IF NOT EXISTS device_hash VARCHAR(64);
