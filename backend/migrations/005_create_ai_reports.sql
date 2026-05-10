-- 005_create_ai_reports.sql
-- Tabel laporan AI yang dihasilkan untuk karyawan

CREATE TABLE IF NOT EXISTS ai_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    generated_by    UUID NOT NULL REFERENCES users(id),
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    prompt          TEXT NOT NULL,
    response        TEXT NOT NULL,
    model_used      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_reports_employee_id ON ai_reports(employee_id);
