-- 006_seed_data.sql
-- Seed data awal: 1 super_admin + 3 karyawan + 7 hari absensi per karyawan
-- Password default semua user: admin123
-- Bcrypt hash di bawah di-generate dengan cost 10 dan sudah diverifikasi
-- terhadap password "admin123" via golang.org/x/crypto/bcrypt.
-- Catatan: ON CONFLICT DO NOTHING agar re-run migration aman.

-- =============================================================================
-- 1. Super Admin
-- =============================================================================
INSERT INTO users (id, company_id, name, email, password_hash, role, department, position, nik, is_active, joined_at)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000001',
    'Super Admin',
    'admin@company.com',
    '$2a$10$l6u0ifdZC7vhmBwk0ytMvule.DUQUC2q9l8KfwPXT2PSza3.R2nYe',
    'super_admin',
    'Manajemen',
    'System Administrator',
    'ADM-0001',
    true,
    CURRENT_DATE - INTERVAL '365 days'
) ON CONFLICT (email) DO NOTHING;

-- =============================================================================
-- 2. Tiga Karyawan dengan Departemen Berbeda
-- =============================================================================
-- Budi Santoso — Teknologi
INSERT INTO users (id, company_id, name, email, password_hash, role, department, position, nik, is_active, joined_at)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '00000000-0000-0000-0000-000000000001',
    'Budi Santoso',
    'budi@company.com',
    '$2a$10$l6u0ifdZC7vhmBwk0ytMvule.DUQUC2q9l8KfwPXT2PSza3.R2nYe',
    'karyawan',
    'Teknologi',
    'Software Engineer',
    'EMP-0001',
    true,
    CURRENT_DATE - INTERVAL '180 days'
) ON CONFLICT (email) DO NOTHING;

-- Siti Nurhaliza — HR
INSERT INTO users (id, company_id, name, email, password_hash, role, department, position, nik, is_active, joined_at)
VALUES (
    '33333333-3333-3333-3333-333333333333',
    '00000000-0000-0000-0000-000000000001',
    'Siti Nurhaliza',
    'siti@company.com',
    '$2a$10$l6u0ifdZC7vhmBwk0ytMvule.DUQUC2q9l8KfwPXT2PSza3.R2nYe',
    'karyawan',
    'HR',
    'HR Officer',
    'EMP-0002',
    true,
    CURRENT_DATE - INTERVAL '120 days'
) ON CONFLICT (email) DO NOTHING;

-- Agus Wijaya — Keuangan
INSERT INTO users (id, company_id, name, email, password_hash, role, department, position, nik, is_active, joined_at)
VALUES (
    '44444444-4444-4444-4444-444444444444',
    '00000000-0000-0000-0000-000000000001',
    'Agus Wijaya',
    'agus@company.com',
    '$2a$10$l6u0ifdZC7vhmBwk0ytMvule.DUQUC2q9l8KfwPXT2PSza3.R2nYe',
    'karyawan',
    'Keuangan',
    'Accountant',
    'EMP-0003',
    true,
    CURRENT_DATE - INTERVAL '90 days'
) ON CONFLICT (email) DO NOTHING;

-- =============================================================================
-- 3. Data Absensi 7 Hari Terakhir per Karyawan
-- Mix status: hadir, terlambat, izin
-- =============================================================================

-- Budi Santoso (Teknologi) — pola: mostly hadir, 1 terlambat, 1 izin
INSERT INTO attendances (user_id, date, clock_in, clock_out, status, notes, ip_address)
VALUES
    ('22222222-2222-2222-2222-222222222222', CURRENT_DATE - 6, (CURRENT_DATE - 6 + TIME '08:05:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 6 + TIME '17:10:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', 'Normal', '192.168.1.10'),
    ('22222222-2222-2222-2222-222222222222', CURRENT_DATE - 5, (CURRENT_DATE - 5 + TIME '08:45:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 5 + TIME '17:30:00') AT TIME ZONE 'Asia/Jakarta', 'terlambat', 'Macet di jalan', '192.168.1.10'),
    ('22222222-2222-2222-2222-222222222222', CURRENT_DATE - 4, (CURRENT_DATE - 4 + TIME '07:55:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 4 + TIME '17:05:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.10'),
    ('22222222-2222-2222-2222-222222222222', CURRENT_DATE - 3, NULL, NULL, 'izin', 'Izin keperluan keluarga', NULL),
    ('22222222-2222-2222-2222-222222222222', CURRENT_DATE - 2, (CURRENT_DATE - 2 + TIME '08:00:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 2 + TIME '17:15:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.10'),
    ('22222222-2222-2222-2222-222222222222', CURRENT_DATE - 1, (CURRENT_DATE - 1 + TIME '08:10:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 1 + TIME '17:20:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.10'),
    ('22222222-2222-2222-2222-222222222222', CURRENT_DATE,     (CURRENT_DATE + TIME '08:02:00')     AT TIME ZONE 'Asia/Jakarta', NULL,                                                                     'hadir', NULL, '192.168.1.10')
ON CONFLICT (user_id, date) DO NOTHING;

-- Siti Nurhaliza (HR) — pola: lebih disiplin, 1 terlambat, 1 izin
INSERT INTO attendances (user_id, date, clock_in, clock_out, status, notes, ip_address)
VALUES
    ('33333333-3333-3333-3333-333333333333', CURRENT_DATE - 6, (CURRENT_DATE - 6 + TIME '07:50:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 6 + TIME '17:00:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.11'),
    ('33333333-3333-3333-3333-333333333333', CURRENT_DATE - 5, (CURRENT_DATE - 5 + TIME '07:55:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 5 + TIME '17:05:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.11'),
    ('33333333-3333-3333-3333-333333333333', CURRENT_DATE - 4, NULL, NULL, 'izin', 'Izin kontrol kesehatan', NULL),
    ('33333333-3333-3333-3333-333333333333', CURRENT_DATE - 3, (CURRENT_DATE - 3 + TIME '08:30:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 3 + TIME '17:15:00') AT TIME ZONE 'Asia/Jakarta', 'terlambat', 'Antar anak ke sekolah', '192.168.1.11'),
    ('33333333-3333-3333-3333-333333333333', CURRENT_DATE - 2, (CURRENT_DATE - 2 + TIME '07:48:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 2 + TIME '17:00:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.11'),
    ('33333333-3333-3333-3333-333333333333', CURRENT_DATE - 1, (CURRENT_DATE - 1 + TIME '07:52:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 1 + TIME '17:10:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.11'),
    ('33333333-3333-3333-3333-333333333333', CURRENT_DATE,     (CURRENT_DATE + TIME '07:58:00')     AT TIME ZONE 'Asia/Jakarta', NULL,                                                                     'hadir', NULL, '192.168.1.11')
ON CONFLICT (user_id, date) DO NOTHING;

-- Agus Wijaya (Keuangan) — pola: beberapa kali terlambat, 1 izin
INSERT INTO attendances (user_id, date, clock_in, clock_out, status, notes, ip_address)
VALUES
    ('44444444-4444-4444-4444-444444444444', CURRENT_DATE - 6, (CURRENT_DATE - 6 + TIME '08:15:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 6 + TIME '17:20:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.12'),
    ('44444444-4444-4444-4444-444444444444', CURRENT_DATE - 5, (CURRENT_DATE - 5 + TIME '08:50:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 5 + TIME '17:30:00') AT TIME ZONE 'Asia/Jakarta', 'terlambat', 'Kendaraan mogok', '192.168.1.12'),
    ('44444444-4444-4444-4444-444444444444', CURRENT_DATE - 4, (CURRENT_DATE - 4 + TIME '08:10:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 4 + TIME '17:15:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.12'),
    ('44444444-4444-4444-4444-444444444444', CURRENT_DATE - 3, (CURRENT_DATE - 3 + TIME '09:05:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 3 + TIME '18:00:00') AT TIME ZONE 'Asia/Jakarta', 'terlambat', 'Rapat di luar', '192.168.1.12'),
    ('44444444-4444-4444-4444-444444444444', CURRENT_DATE - 2, NULL, NULL, 'izin', 'Izin acara keluarga', NULL),
    ('44444444-4444-4444-4444-444444444444', CURRENT_DATE - 1, (CURRENT_DATE - 1 + TIME '08:05:00') AT TIME ZONE 'Asia/Jakarta', (CURRENT_DATE - 1 + TIME '17:10:00') AT TIME ZONE 'Asia/Jakarta', 'hadir', NULL, '192.168.1.12'),
    ('44444444-4444-4444-4444-444444444444', CURRENT_DATE,     (CURRENT_DATE + TIME '08:20:00')     AT TIME ZONE 'Asia/Jakarta', NULL,                                                                     'hadir', NULL, '192.168.1.12')
ON CONFLICT (user_id, date) DO NOTHING;
