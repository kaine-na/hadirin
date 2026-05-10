# Brainstorming Teknis — SaaS Karyawan
**Tanggal:** 2026-05-10
**Author:** Forge (AI Coding Agent)
**Stack:** SvelteKit + TailwindCSS (frontend) | Go + Gin/Chi + PostgreSQL (backend)

---

## 1. Technical Debt & Quick Wins

### Quick Wins (3 item — bisa langsung dikerjakan)

#### QW-1: Rate Limiting di Login Endpoint
**File:** `backend/cmd/server/main.go`
**Masalah:** Endpoint `/api/auth/login` tidak ada rate limiting — rentan brute force.
**Fix:** Tambahkan middleware `go-chi/httprate` (sudah pakai chi, tinggal tambah dependency):
```go
import "github.com/go-chi/httprate"

r.With(httprate.LimitByIP(5, 1*time.Minute)).Post("/login", authHandler.Login)
```
**Effort:** S (< 1 jam) | **Impact:** HIGH (security)

#### QW-2: Validasi period_start <= period_end di AI Analyze
**File:** `backend/internal/ai/service.go` baris 25-33
**Masalah:** Tidak ada validasi bahwa `period_start` harus sebelum `period_end`. Query tetap jalan dan hasilkan data kosong tanpa error yang jelas.
**Fix:**
```go
if !periodStart.Before(periodEnd) && !periodStart.Equal(periodEnd) {
    return nil, errors.New("period_start harus sebelum atau sama dengan period_end")
}
```
**Effort:** S (< 30 menit) | **Impact:** MEDIUM (UX + data integrity)

#### QW-3: Proper Pagination di Frontend Employees
**File:** `frontend/src/routes/(app)/employees/+page.svelte`
**Masalah:** `employeesApi.list({ page_size: 200 })` hardcoded — tapi backend max 100. Inkonsistensi ini menyebabkan data terpotong diam-diam.
**Fix:** Implementasi cursor/offset pagination yang proper dengan tombol "Load More" atau paginator. Backend sudah support `page` dan `page_size`, tinggal frontend yang diupdate.
**Effort:** S-M (2-4 jam) | **Impact:** MEDIUM (correctness + performance)

---

### Technical Debt (2 item — perlu diatasi sebelum scale)

#### TD-1: JWT Logout Tidak Invalidate Token (Token Blacklist)
**File:** `backend/internal/auth/handler.go` — `Logout()`
**Masalah:** Logout hanya return 200 OK tanpa invalidate token. Token tetap valid sampai expired (default 24 jam). Untuk aplikasi HR yang menyimpan data sensitif karyawan, ini risiko nyata — jika token bocor, attacker punya akses 24 jam penuh.
**Solusi yang direkomendasikan:**
- Tambahkan Redis sebagai token blacklist store
- Saat logout, simpan `jti` (JWT ID) ke Redis dengan TTL = sisa waktu token
- Di `ValidateToken`, cek apakah `jti` ada di blacklist
- Alternatif lebih ringan: kurangi `JWT_EXPIRY_HOURS` ke 1 jam + implementasi refresh token

**Effort:** M (1-2 hari) | **Impact:** HIGH (security, wajib sebelum production)

#### TD-2: File Storage Lokal — Tidak Scalable
**File:** `backend/internal/document/service.go`, `backend/internal/employee/service.go`
**Masalah:** File upload (dokumen + foto karyawan) disimpan di filesystem lokal (`./uploads`). Ini berarti:
- Tidak bisa horizontal scaling (file hanya ada di satu server)
- Tidak ada backup otomatis
- Docker container restart = file hilang (kecuali ada volume mount)
- Multi-tenant nanti akan campur aduk file antar perusahaan

**Solusi yang direkomendasikan:**
- Migrasi ke object storage: MinIO (self-hosted, S3-compatible) atau Cloudflare R2 (murah, S3-compatible)
- Buat abstraksi `StorageProvider` interface di Go:
```go
type StorageProvider interface {
    Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error)
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    GetURL(key string) string
}
```
- Implementasi `LocalStorage` (existing) dan `S3Storage` (baru) — bisa switch via env var

**Effort:** L (3-5 hari) | **Impact:** CRITICAL untuk scale

---

## 2. Ide Fitur dari Perspektif Teknis

### Fitur 1: Smart Leave Management (Manajemen Cuti Cerdas)

**Nama fitur:** Smart Leave Management
**Deskripsi teknis:**
Sistem pengajuan dan approval cuti yang terintegrasi dengan data absensi yang sudah ada. Karyawan bisa mengajukan cuti (tahunan, sakit, izin khusus), manager/HR approve/reject, dan sistem otomatis update status absensi. AI bisa memberikan rekomendasi apakah cuti layak disetujui berdasarkan pola absensi historis.

Implementasi di backend: tabel `leave_requests` baru dengan state machine (pending → approved/rejected). Trigger PostgreSQL atau logic di service untuk auto-create attendance record dengan status "izin" saat cuti disetujui.

**Endpoint baru yang dibutuhkan:**
```
POST   /api/leaves              — ajukan cuti
GET    /api/leaves              — list cuti (filter: status, user, periode)
GET    /api/leaves/me           — cuti saya
GET    /api/leaves/{id}         — detail cuti
PUT    /api/leaves/{id}/approve — approve (HR/Manager)
PUT    /api/leaves/{id}/reject  — reject dengan alasan
GET    /api/leaves/balance/{user_id} — sisa jatah cuti
```

**Komponen frontend baru:**
- `LeaveRequestForm.svelte` — form pengajuan dengan date picker
- `LeaveApprovalCard.svelte` — card untuk HR approve/reject
- `LeaveBalanceWidget.svelte` — widget sisa cuti di dashboard
- Route: `/leaves` (list), `/leaves/request` (form baru)

**Estimasi effort:** M (3-5 hari backend + 2-3 hari frontend)

**Dependencies:**
- Tidak butuh library baru
- Opsional: `github.com/robfig/cron/v3` untuk auto-reset jatah cuti tahunan tiap tahun

---

### Fitur 2: AI Chat HR Assistant (Chatbot HR)

**Nama fitur:** AI Chat HR Assistant
**Deskripsi teknis:**
Chatbot berbasis LLM yang bisa menjawab pertanyaan HR secara natural language. Memanfaatkan LLM API yang sudah ada (custom OpenAI-compatible). Karyawan bisa tanya "Berapa sisa cuti saya?", "Kapan gajian bulan ini?", "Bagaimana cara mengajukan izin?". HR bisa tanya "Siapa yang paling sering terlambat bulan ini?", "Buat draft surat peringatan untuk karyawan X".

Implementasi: RAG sederhana — ambil data relevan dari DB berdasarkan intent user, inject ke prompt LLM. Tidak perlu vector database untuk MVP — cukup keyword matching + SQL query yang sudah ada.

**Endpoint baru yang dibutuhkan:**
```
POST /api/ai/chat              — kirim pesan, terima respons streaming
GET  /api/ai/chat/history      — riwayat chat session
DELETE /api/ai/chat/history    — hapus riwayat
```

**Komponen frontend baru:**
- `ChatWindow.svelte` — UI chat dengan streaming response (SSE)
- `ChatMessage.svelte` — bubble pesan user/AI
- `ChatInput.svelte` — input dengan submit on Enter
- Route: `/hr-ai/chat` (tab baru di halaman HR AI yang sudah ada)

**Estimasi effort:** M (2-3 hari backend + 2 hari frontend)

**Dependencies:**
- Backend: tidak ada library baru (LLM client sudah ada)
- Frontend: tidak ada library baru (SSE native di browser)
- Opsional: `github.com/gorilla/websocket` jika mau WebSocket instead of SSE

---

### Fitur 3: Notifikasi & Reminder Otomatis

**Nama fitur:** Smart Notification System
**Deskripsi teknis:**
Sistem notifikasi in-app + email untuk event penting: reminder clock-in (jam 08:00 jika belum absen), notifikasi approval dokumen, reminder cuti yang akan habis, alert karyawan yang sering terlambat. Backend menggunakan cron job (goroutine dengan ticker) yang query DB dan kirim notifikasi.

Implementasi: tabel `notifications` di PostgreSQL, endpoint SSE untuk real-time push ke frontend, cron goroutine di background. Email via SMTP (bisa pakai `net/smtp` stdlib atau `gomail`).

**Endpoint baru yang dibutuhkan:**
```
GET  /api/notifications         — list notifikasi user (unread first)
PUT  /api/notifications/{id}/read — tandai sudah dibaca
PUT  /api/notifications/read-all  — tandai semua sudah dibaca
GET  /api/notifications/stream    — SSE endpoint untuk real-time
DELETE /api/notifications/{id}    — hapus notifikasi
```

**Komponen frontend baru:**
- `NotificationBell.svelte` — icon lonceng di navbar dengan badge count
- `NotificationDropdown.svelte` — dropdown list notifikasi
- `NotificationItem.svelte` — item individual dengan icon dan timestamp
- Store: `notifications.svelte.ts` — Svelte 5 runes store untuk state

**Estimasi effort:** M (3-4 hari backend + 2 hari frontend)

**Dependencies:**
- Backend: `gopkg.in/gomail.v2` untuk email (atau `net/smtp` stdlib)
- Tidak butuh library baru untuk SSE dan cron (pakai `time.Ticker`)

---

### Fitur 4: Laporan & Analytics Dashboard (Export PDF + Charts)

**Nama fitur:** HR Analytics & Reporting
**Deskripsi teknis:**
Dashboard analytics yang lebih kaya dengan chart interaktif (kehadiran per departemen, tren keterlambatan, distribusi status absensi) dan kemampuan export laporan ke PDF. Memanfaatkan data yang sudah ada di PostgreSQL dengan query agregasi. AI bisa generate executive summary dari data laporan.

Implementasi backend: endpoint agregasi baru yang return data siap chart. PDF generation menggunakan `go-pdf/fpdf` atau `jung-kurt/gofpdf`. Frontend menggunakan Chart.js atau Recharts (sudah umum di ekosistem SvelteKit).

**Endpoint baru yang dibutuhkan:**
```
GET /api/reports/attendance/summary    — ringkasan kehadiran (filter: dept, periode)
GET /api/reports/attendance/trend      — tren harian/mingguan/bulanan
GET /api/reports/employees/department  — distribusi per departemen
GET /api/reports/export/pdf            — export laporan ke PDF
GET /api/reports/export/excel          — export ke Excel (opsional)
POST /api/ai/report-summary            — AI generate executive summary dari data
```

**Komponen frontend baru:**
- `AttendanceChart.svelte` — bar/line chart kehadiran
- `DepartmentPieChart.svelte` — pie chart distribusi departemen
- `ReportFilter.svelte` — filter periode, departemen, status
- `ExportButton.svelte` — tombol export PDF/CSV
- Route: `/reports` (halaman baru)

**Estimasi effort:** L (4-6 hari backend + 3-4 hari frontend)

**Dependencies:**
- Backend: `github.com/jung-kurt/gofpdf` atau `github.com/signintech/gopdf` untuk PDF
- Frontend: `chart.js` + `svelte-chartjs` wrapper (ringan, ~60KB)

---

### Fitur 5: Payroll Calculator (Kalkulasi Gaji Otomatis)

**Nama fitur:** Payroll Calculator
**Deskripsi teknis:**
Modul kalkulasi gaji sederhana yang mengintegrasikan data absensi dengan komponen gaji. HR input gaji pokok, tunjangan, dan aturan potongan (per hari alpha, per jam terlambat). Sistem otomatis hitung gaji bersih berdasarkan data absensi bulan tersebut. AI bisa bantu generate slip gaji dalam format yang readable.

Implementasi: tabel `salary_configs` (gaji pokok per karyawan) dan `payroll_records` (hasil kalkulasi per bulan). Logic kalkulasi di service layer — murni Go, tidak butuh library eksternal. AI generate narasi slip gaji dari data numerik.

**Endpoint baru yang dibutuhkan:**
```
POST /api/payroll/config/{employee_id}  — set konfigurasi gaji
GET  /api/payroll/config/{employee_id}  — get konfigurasi gaji
POST /api/payroll/calculate             — hitung gaji (periode, employee_ids)
GET  /api/payroll/records               — list hasil kalkulasi
GET  /api/payroll/records/{id}          — detail slip gaji
GET  /api/payroll/records/{id}/pdf      — export slip gaji PDF
POST /api/ai/payroll-summary/{id}       — AI generate narasi slip gaji
```

**Komponen frontend baru:**
- `SalaryConfigForm.svelte` — form input gaji pokok + tunjangan
- `PayrollCalculator.svelte` — UI kalkulasi dengan preview
- `PayslipCard.svelte` — tampilan slip gaji
- Route: `/payroll` (halaman baru, akses HR Admin only)

**Estimasi effort:** L (5-7 hari backend + 3-4 hari frontend)

**Dependencies:**
- Tidak butuh library baru untuk kalkulasi
- Reuse `gofpdf` dari fitur Reports untuk export slip gaji PDF

---

## 3. Arsitektur untuk Scale (1000+ Perusahaan Multi-Tenant)

### Rekomendasi 1: Row-Level Multi-Tenancy di PostgreSQL

**Masalah saat ini:** Arsitektur sekarang single-tenant — tidak ada kolom `company_id` di tabel utama (users, attendances, documents). Jika langsung deploy untuk 1000 perusahaan di satu database, data antar perusahaan bisa bocor jika ada bug di query.

**Rekomendasi konkret:**
1. Tambahkan kolom `company_id UUID NOT NULL` ke semua tabel utama
2. Buat PostgreSQL Row-Level Security (RLS) policy:
```sql
-- Enable RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE attendances ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

-- Policy: user hanya bisa akses data company mereka
CREATE POLICY tenant_isolation ON users
    USING (company_id = current_setting('app.current_company_id')::uuid);
```
3. Di Go middleware, set `SET LOCAL app.current_company_id = '<id>'` di awal setiap request
4. JWT claims sudah ada `company_id` — tinggal dipakai di middleware

**Benefit:** Isolasi data di level database, bukan hanya aplikasi. Bahkan jika ada bug di query (lupa WHERE company_id), RLS tetap block akses.

**Effort migrasi:** L (perlu migration script + update semua query + testing menyeluruh)

---

### Rekomendasi 2: Object Storage + CDN untuk File

**Masalah saat ini:** File disimpan di filesystem lokal. Tidak bisa horizontal scaling, tidak ada backup, tidak ada CDN.

**Rekomendasi konkret:**
1. Implementasi `StorageProvider` interface (lihat TD-2 di atas)
2. Gunakan **Cloudflare R2** (S3-compatible, gratis egress, murah untuk Indonesia):
   - Bucket per tenant: `company-{company_id}/documents/`, `company-{company_id}/photos/`
   - Atau single bucket dengan prefix: `{company_id}/{type}/{filename}`
3. Generate **presigned URL** untuk download (bukan serve langsung dari backend):
```go
// Backend hanya generate URL, tidak proxy file
func (h *DocumentHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
    url, err := h.storage.GetPresignedURL(ctx, doc.StorageKey, 15*time.Minute)
    // Return URL ke frontend, frontend redirect langsung ke R2/CDN
}
```
4. Tambahkan Cloudflare CDN di depan R2 untuk cache foto profil dan dokumen publik

**Benefit:** Horizontal scaling backend tanpa state, backup otomatis, CDN untuk performa, biaya storage lebih murah dari VPS disk.

**Effort migrasi:** M-L (2-3 hari implementasi + migration script untuk file existing)

---

### Rekomendasi 3: Caching Layer + Background Job Queue

**Masalah saat ini:** Setiap request AI analyze langsung hit LLM API (latency tinggi, biaya mahal). Tidak ada caching untuk data yang sering diakses (dashboard stats, employee list).

**Rekomendasi konkret:**

**A. Redis untuk Caching:**
```go
// Cache dashboard stats 5 menit
func (s *Service) GetDashboardStats(ctx context.Context, companyID string) (*Stats, error) {
    cacheKey := fmt.Sprintf("stats:%s", companyID)
    if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
        // Return dari cache
    }
    // Query DB, simpan ke cache 5 menit
    s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
}
```

**B. Background Job Queue untuk AI:**
Saat ini AI analyze adalah synchronous — user tunggu 10-30 detik. Untuk scale, ubah ke async:
```
POST /api/ai/analyze/{employee_id}
→ Return: { "job_id": "uuid", "status": "queued" }

GET /api/ai/jobs/{job_id}
→ Return: { "status": "processing|done|failed", "result": {...} }
```
Implementasi dengan **Asynq** (Redis-backed job queue untuk Go) atau simple PostgreSQL-based queue.

**C. Connection Pooling yang Proper:**
Saat ini `pgxpool` sudah dipakai (bagus), tapi untuk 1000 tenant perlu tuning:
```go
config.MaxConns = 50          // Sesuaikan dengan RAM PostgreSQL
config.MinConns = 5
config.MaxConnLifetime = 1 * time.Hour
config.MaxConnIdleTime = 30 * time.Minute
```
Pertimbangkan **PgBouncer** di depan PostgreSQL untuk connection pooling di level infrastruktur.

**Benefit:** Latency dashboard turun drastis, AI analyze tidak block user, PostgreSQL tidak kewalahan dengan 1000 concurrent connections.

**Effort implementasi:** M-L (Redis setup + Asynq integration + frontend polling untuk async jobs)

---

## 4. Nama dari Perspektif Teknis/Developer

### Nama 1: **Hadirku**
- **Domain:** `hadirku.id` atau `hadirku.co.id`
- **Alasan teknis:** Tidak ada konflik dengan package Go atau npm. Mudah diketik, tidak ambigu. Kata "hadir" langsung relevan dengan fitur utama (absensi). Suffix "-ku" memberi kesan personal/SaaS.
- **Cek konflik:** `npm search hadirku` → tidak ada. `pkg.go.dev/hadirku` → tidak ada.
- **Skor developer-friendliness:** 9/10 — pendek, fonetik jelas, mudah diingat.

### Nama 2: **Kerjain**
- **Domain:** `kerjain.id` atau `kerjain.co.id`
- **Alasan teknis:** Tidak ada konflik package. "Kerja" = work, "-in" = suffix informal Indonesia yang populer di startup (mirip Tokopedia, Gojek). Mudah diketik di terminal (`kerjain-api`, `kerjain-frontend`). Cocok sebagai nama repo, Docker image, dan subdomain.
- **Cek konflik:** Tidak ada library populer bernama kerjain.
- **Skor developer-friendliness:** 8/10 — sedikit lebih panjang tapi tetap catchy.

### Nama 3: **Rekap.id**
- **Domain:** `rekap.id` (langsung pakai .id sebagai bagian nama)
- **Alasan teknis:** "Rekap" adalah kata yang sudah familiar di konteks HR Indonesia (rekap absensi, rekap gaji). Sangat pendek — ideal untuk CLI tools, API subdomain (`api.rekap.id`), dan Docker image names (`rekap/backend`). Tidak ada konflik dengan package populer.
- **Cek konflik:** Tidak ada npm package atau Go module bernama rekap.
- **Skor developer-friendliness:** 9/10 — paling pendek, paling mudah diketik, domain .id tersedia.

---

## Ringkasan Prioritas

| Item | Kategori | Effort | Impact | Prioritas |
|------|----------|--------|--------|-----------|
| Rate limiting login | Quick Win | S | HIGH | 🔴 Segera |
| Validasi period AI | Quick Win | S | MEDIUM | 🟡 Sprint ini |
| Proper pagination | Quick Win | S-M | MEDIUM | 🟡 Sprint ini |
| JWT token blacklist | Tech Debt | M | HIGH | 🔴 Sebelum launch |
| File storage migration | Tech Debt | L | CRITICAL | 🔴 Sebelum scale |
| Smart Leave Management | Fitur | M | HIGH | 🟢 Sprint 1 |
| AI Chat HR Assistant | Fitur | M | HIGH | 🟢 Sprint 1 |
| Notifikasi Otomatis | Fitur | M | MEDIUM | 🟡 Sprint 2 |
| HR Analytics & PDF | Fitur | L | MEDIUM | 🟡 Sprint 2 |
| Payroll Calculator | Fitur | L | HIGH | 🟢 Sprint 3 |
| Multi-tenant RLS | Arsitektur | L | CRITICAL | 🔴 Sebelum scale |
| Object Storage | Arsitektur | M-L | CRITICAL | 🔴 Sebelum scale |
| Redis + Job Queue | Arsitektur | M-L | HIGH | 🟡 Saat 100+ tenant |

---

*Dokumen ini dibuat oleh Forge (AI Coding Agent) berdasarkan review codebase aktual.*
*Stack: SvelteKit + TailwindCSS | Go + Chi + PostgreSQL | pgx/v5*
