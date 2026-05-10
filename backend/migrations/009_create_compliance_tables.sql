-- Migration 009: Compliance Engine Tables
-- Tabel untuk BPJS, PPh21, THR, dan checklist kepatuhan

-- compliance_rules: aturan regulasi yang berlaku
CREATE TABLE IF NOT EXISTS compliance_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(50) NOT NULL UNIQUE,  -- e.g. BPJS_KESEHATAN, PPH21_TER, THR
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    rule_type   VARCHAR(50) NOT NULL,         -- bpjs, pph21, thr, checklist
    parameters  JSONB NOT NULL DEFAULT '{}',  -- parameter tarif, batas, dll
    effective_from DATE NOT NULL,
    effective_to   DATE,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- compliance_records: hasil kalkulasi per karyawan per periode
CREATE TABLE IF NOT EXISTS compliance_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period          VARCHAR(7) NOT NULL,  -- format YYYY-MM
    record_type     VARCHAR(50) NOT NULL, -- bpjs, pph21, thr
    gross_salary    BIGINT NOT NULL DEFAULT 0,
    -- BPJS fields
    bpjs_kes_company    BIGINT DEFAULT 0,
    bpjs_kes_employee   BIGINT DEFAULT 0,
    bpjs_jht_company    BIGINT DEFAULT 0,
    bpjs_jht_employee   BIGINT DEFAULT 0,
    bpjs_jp_company     BIGINT DEFAULT 0,
    bpjs_jp_employee    BIGINT DEFAULT 0,
    bpjs_jkk_company    BIGINT DEFAULT 0,
    bpjs_jkm_company    BIGINT DEFAULT 0,
    -- PPh21 fields
    pph21_ter_category  VARCHAR(5) DEFAULT '',   -- A, B, C
    pph21_ter_rate      NUMERIC(5,2) DEFAULT 0,  -- persentase
    pph21_amount        BIGINT DEFAULT 0,
    pph21_ytd_gross     BIGINT DEFAULT 0,        -- year-to-date gross
    pph21_ytd_tax       BIGINT DEFAULT 0,        -- year-to-date tax
    -- THR fields
    thr_amount          BIGINT DEFAULT 0,
    thr_religion        VARCHAR(50) DEFAULT '',
    thr_holiday_date    DATE,
    -- metadata
    calculation_details JSONB DEFAULT '{}',
    calculated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, period, record_type)
);

-- compliance_checklist: checklist kepatuhan per bulan
CREATE TABLE IF NOT EXISTS compliance_checklist (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period      VARCHAR(7) NOT NULL,  -- format YYYY-MM
    item_code   VARCHAR(100) NOT NULL,
    title       VARCHAR(200) NOT NULL,
    description TEXT,
    deadline    DATE NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, done, overdue
    done_at     TIMESTAMPTZ,
    done_by     UUID REFERENCES users(id),
    notified_h3 BOOLEAN NOT NULL DEFAULT false,  -- sudah kirim notif H-3?
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(period, item_code)
);

-- Index untuk performa query
CREATE INDEX IF NOT EXISTS idx_compliance_records_user_period ON compliance_records(user_id, period);
CREATE INDEX IF NOT EXISTS idx_compliance_records_period_type ON compliance_records(period, record_type);
CREATE INDEX IF NOT EXISTS idx_compliance_checklist_period ON compliance_checklist(period);
CREATE INDEX IF NOT EXISTS idx_compliance_checklist_status ON compliance_checklist(status);
CREATE INDEX IF NOT EXISTS idx_compliance_checklist_deadline ON compliance_checklist(deadline);

-- Seed aturan compliance default
INSERT INTO compliance_rules (code, name, description, rule_type, parameters, effective_from) VALUES
(
    'BPJS_KESEHATAN',
    'BPJS Kesehatan',
    'Iuran BPJS Kesehatan sesuai Perpres 64/2020',
    'bpjs',
    '{
        "company_rate": 0.04,
        "employee_rate": 0.01,
        "max_salary": 12000000
    }',
    '2020-07-01'
),
(
    'BPJS_TK_JHT',
    'BPJS Ketenagakerjaan - JHT',
    'Jaminan Hari Tua',
    'bpjs',
    '{
        "company_rate": 0.037,
        "employee_rate": 0.02
    }',
    '2015-07-01'
),
(
    'BPJS_TK_JP',
    'BPJS Ketenagakerjaan - JP',
    'Jaminan Pensiun',
    'bpjs',
    '{
        "company_rate": 0.02,
        "employee_rate": 0.01,
        "max_salary": 9559600
    }',
    '2015-07-01'
),
(
    'BPJS_TK_JKK',
    'BPJS Ketenagakerjaan - JKK',
    'Jaminan Kecelakaan Kerja (risiko standar)',
    'bpjs',
    '{
        "company_rate": 0.0024
    }',
    '2015-07-01'
),
(
    'BPJS_TK_JKM',
    'BPJS Ketenagakerjaan - JKM',
    'Jaminan Kematian',
    'bpjs',
    '{
        "company_rate": 0.003
    }',
    '2015-07-01'
),
(
    'PPH21_TER',
    'PPh 21 Metode TER',
    'Pajak Penghasilan Pasal 21 dengan Tarif Efektif Rata-rata sesuai PMK 168/2023',
    'pph21',
    '{
        "method": "TER",
        "regulation": "PMK 168/2023"
    }',
    '2024-01-01'
),
(
    'THR',
    'Tunjangan Hari Raya',
    'THR sesuai PP 36/2021',
    'thr',
    '{
        "full_service_months": 12,
        "religions": ["islam", "kristen", "katolik", "hindu", "buddha", "konghucu"]
    }',
    '2021-03-02'
)
ON CONFLICT (code) DO NOTHING;
