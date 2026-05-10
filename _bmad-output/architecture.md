# Architecture — Hadir: Platform Manajemen Karyawan Digital
**Versi:** 2.0  
**Tanggal:** 2026-05-10  
**Stack:** Go (Backend) + SvelteKit (Frontend) + PostgreSQL

---

## 1. High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Browser (SvelteKit)                   │
│  /login  /dashboard  /attendance  /documents  /hr-ai    │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTP/REST (JSON)
                       ▼
┌─────────────────────────────────────────────────────────┐
│                   Go HTTP Server                         │
│  Chi Router + JWT Middleware + CORS                      │
│                                                          │
│  /api/auth/*    /api/employees/*   /api/attendance/*    │
│  /api/documents/*    /api/ai/*                          │
└──────────┬──────────────────────────────────────────────┘
           │                              │
           ▼                              ▼
┌──────────────────┐          ┌──────────────────────────┐
│   PostgreSQL     │          │  LLM API (OpenAI-compat) │
│   (pgx/v5)       │          │  http://43.133.61.163    │
└──────────────────┘          └──────────────────────────┘
           │
           ▼
┌──────────────────┐
│  Local Filesystem│
│  ./uploads/      │
└──────────────────┘
```

---

## 2. Backend Structure (Go)

```
saas-karyawan/
├── cmd/
│   └── server/
│       └── main.go              # Entry point, setup router, start server
├── internal/
│   ├── auth/
│   │   ├── handler.go           # HTTP handlers: login, logout, me
│   │   ├── service.go           # Business logic: validate credentials, issue JWT
│   │   ├── middleware.go        # JWT validation middleware
│   │   └── model.go             # LoginRequest, TokenResponse structs
│   ├── employee/
│   │   ├── handler.go           # HTTP handlers: CRUD employees
│   │   ├── service.go           # Business logic: create, update, soft delete
│   │   ├── repository.go        # DB queries: find, list, save employees
│   │   └── model.go             # Employee struct, CreateEmployeeRequest
│   ├── attendance/
│   │   ├── handler.go           # HTTP handlers: clock-in, clock-out, rekap, export
│   │   ├── service.go           # Business logic: status calculation, CSV generation
│   │   ├── repository.go        # DB queries: attendance records
│   │   └── model.go             # Attendance struct, ClockInRequest
│   ├── document/
│   │   ├── handler.go           # HTTP handlers: upload, list, get, delete
│   │   ├── service.go           # Business logic: file validation, versioning
│   │   ├── repository.go        # DB queries: documents, comments
│   │   └── model.go             # Document struct, UploadRequest
│   ├── ai/
│   │   ├── handler.go           # HTTP handlers: analyze, get reports
│   │   ├── service.go           # Business logic: build prompt, call LLM, save report
│   │   ├── client.go            # OpenAI-compatible HTTP client
│   │   └── model.go             # AIReport struct, AnalyzeRequest
│   └── database/
│       ├── postgres.go          # Connection pool setup (pgx/v5)
│       └── migrate.go           # Run SQL migrations on startup
├── pkg/
│   ├── config/
│   │   └── config.go            # Load .env, Config struct
│   └── response/
│       └── response.go          # Standard JSON response helpers
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_attendances.sql
│   ├── 003_create_documents.sql
│   ├── 004_create_document_comments.sql
│   └── 005_create_ai_reports.sql
├── uploads/                     # File storage (gitignored)
├── go.mod
├── go.sum
└── .env.example
```

### 2.1 Dependency Stack (Go)
| Package | Versi | Kegunaan |
|---------|-------|----------|
| `github.com/go-chi/chi/v5` | v5.x | HTTP router |
| `github.com/jackc/pgx/v5` | v5.x | PostgreSQL driver |
| `github.com/golang-jwt/jwt/v5` | v5.x | JWT auth |
| `golang.org/x/crypto` | latest | bcrypt password hashing |
| `github.com/joho/godotenv` | v1.x | Load .env file |

### 2.2 Standard API Response
```go
// pkg/response/response.go
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Contoh sukses:
// {"success": true, "message": "Login berhasil", "data": {"token": "..."}}

// Contoh error:
// {"success": false, "message": "Email atau password salah"}
```

### 2.3 JWT Middleware
```go
// internal/auth/middleware.go
// Ekstrak Bearer token dari header Authorization
// Validasi signature dan expiry
// Inject user claims ke request context
// Cek role untuk endpoint yang membutuhkan role tertentu
```

### 2.4 Config (.env.example)
```
DATABASE_URL=postgres://user:password@localhost:5432/saas_karyawan
JWT_SECRET=your-secret-key-here
JWT_EXPIRY_HOURS=24
UPLOAD_DIR=./uploads
MAX_FILE_SIZE_MB=10
LLM_BASE_URL=http://43.133.61.163:8787/v1
LLM_API_KEY=sk-pool
LLM_MODEL=claude-sonnet-4.6
LLM_TIMEOUT_SECONDS=60
PORT=8080
CORS_ORIGINS=http://localhost:5173
```

---

## 3. Frontend Structure (SvelteKit)

```
frontend/
├── src/
│   ├── routes/
│   │   ├── +layout.svelte           # Root layout, auth check
│   │   ├── +layout.ts               # Load user session
│   │   ├── login/
│   │   │   └── +page.svelte         # Login form
│   │   ├── dashboard/
│   │   │   └── +page.svelte         # Overview stats
│   │   ├── attendance/
│   │   │   ├── +page.svelte         # Clock in/out + rekap personal
│   │   │   └── manage/
│   │   │       └── +page.svelte     # HR: semua absensi + override
│   │   ├── documents/
│   │   │   ├── +page.svelte         # Daftar dokumen
│   │   │   ├── upload/
│   │   │   │   └── +page.svelte     # Form upload
│   │   │   └── [id]/
│   │   │       └── +page.svelte     # Detail dokumen + komentar
│   │   └── hr-ai/
│   │       ├── +page.svelte         # AI dashboard: pilih karyawan
│   │       └── [employee_id]/
│   │           └── +page.svelte     # Laporan AI + riwayat
│   ├── lib/
│   │   ├── components/
│   │   │   ├── Navbar.svelte
│   │   │   ├── Sidebar.svelte
│   │   │   ├── Button.svelte
│   │   │   ├── Table.svelte
│   │   │   ├── Modal.svelte
│   │   │   ├── FileUpload.svelte
│   │   │   └── LoadingSpinner.svelte
│   │   ├── stores/
│   │   │   ├── auth.ts              # User session store (Svelte 5 runes)
│   │   │   └── toast.ts             # Notification store
│   │   ├── api/
│   │   │   ├── client.ts            # Fetch wrapper dengan auth header
│   │   │   ├── auth.ts              # Auth API calls
│   │   │   ├── employees.ts         # Employee API calls
│   │   │   ├── attendance.ts        # Attendance API calls
│   │   │   ├── documents.ts         # Document API calls
│   │   │   └── ai.ts                # AI API calls
│   │   └── types/
│   │       └── index.ts             # TypeScript interfaces
│   └── app.html
├── static/
├── package.json
├── svelte.config.js
├── vite.config.ts
└── tsconfig.json
```

### 3.1 Auth Store (Svelte 5 Runes)
```typescript
// lib/stores/auth.ts
let user = $state<User | null>(null);
let token = $state<string | null>(null);

export function setAuth(u: User, t: string) { ... }
export function clearAuth() { ... }
export function isLoggedIn() { return token !== null; }
export function hasRole(role: string) { return user?.role === role; }
```

### 3.2 API Client
```typescript
// lib/api/client.ts
// Wrapper fetch yang otomatis:
// - Tambah Authorization: Bearer {token} header
// - Parse JSON response
// - Handle 401 → redirect ke /login
// - Handle error response dari API
```

---

## 4. Database Schema (PostgreSQL)

### 4.1 Tabel users
```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL,                    -- multi-tenant
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
CREATE INDEX idx_users_company_id ON users(company_id);
CREATE INDEX idx_users_email ON users(email);
```

### 4.2 Tabel attendances
```sql
CREATE TABLE attendances (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    date        DATE NOT NULL,
    clock_in    TIMESTAMPTZ,
    clock_out   TIMESTAMPTZ,
    status      VARCHAR(20) NOT NULL DEFAULT 'hadir',
                -- 'hadir' | 'terlambat' | 'izin' | 'sakit' | 'alpha'
    notes       TEXT,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    created_by  UUID REFERENCES users(id),        -- jika diinput HR
    updated_by  UUID REFERENCES users(id),        -- jika di-override HR
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, date)
);
CREATE INDEX idx_attendances_user_id ON attendances(user_id);
CREATE INDEX idx_attendances_date ON attendances(date);
```

### 4.3 Tabel documents
```sql
CREATE TABLE documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    category    VARCHAR(100) NOT NULL,
    file_path   VARCHAR(500) NOT NULL,
    file_name   VARCHAR(255) NOT NULL,
    file_size   BIGINT NOT NULL,
    mime_type   VARCHAR(100) NOT NULL,
    version     INTEGER NOT NULL DEFAULT 1,
    parent_id   UUID REFERENCES documents(id),   -- untuk versioning
    doc_date    DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_documents_user_id ON documents(user_id);
CREATE INDEX idx_documents_parent_id ON documents(parent_id);
```

### 4.4 Tabel document_comments
```sql
CREATE TABLE document_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_doc_comments_document_id ON document_comments(document_id);
```

### 4.5 Tabel ai_reports
```sql
CREATE TABLE ai_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES users(id),
    generated_by    UUID NOT NULL REFERENCES users(id),
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    prompt          TEXT NOT NULL,
    response        TEXT NOT NULL,
    model_used      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ai_reports_employee_id ON ai_reports(employee_id);
```

### 4.6 Tabel leave_requests
```sql
CREATE TABLE leave_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    leave_type      VARCHAR(50) NOT NULL,
                    -- 'tahunan' | 'sakit' | 'izin_khusus' | 'melahirkan'
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    total_days      INTEGER NOT NULL,
    reason          TEXT,
    status          VARCHAR(30) NOT NULL DEFAULT 'pending',
                    -- 'pending' | 'approved_manager' | 'approved_hr' | 'rejected' | 'cancelled'
    manager_id      UUID REFERENCES users(id),
    manager_note    TEXT,
    manager_at      TIMESTAMPTZ,
    hr_id           UUID REFERENCES users(id),
    hr_note         TEXT,
    hr_at           TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_leave_requests_user_id ON leave_requests(user_id);
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
CREATE INDEX idx_leave_requests_dates ON leave_requests(start_date, end_date);
```

### 4.7 Tabel leave_balances
```sql
CREATE TABLE leave_balances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    year            INTEGER NOT NULL,
    leave_type      VARCHAR(50) NOT NULL,
    total_days      INTEGER NOT NULL DEFAULT 0,
    used_days       INTEGER NOT NULL DEFAULT 0,
    remaining_days  INTEGER GENERATED ALWAYS AS (total_days - used_days) STORED,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, year, leave_type)
);
CREATE INDEX idx_leave_balances_user_year ON leave_balances(user_id, year);
```

### 4.8 Tabel notifications
```sql
CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    type            VARCHAR(100) NOT NULL,
                    -- 'clock_in_reminder' | 'leave_request_submitted' | 'leave_approved' | etc.
    title           VARCHAR(255) NOT NULL,
    body            TEXT NOT NULL,
    data            JSONB,                              -- payload tambahan (link, id referensi)
    is_read         BOOLEAN NOT NULL DEFAULT false,
    sent_email      BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(user_id, is_read);
```

### 4.9 Tabel notification_preferences
```sql
CREATE TABLE notification_preferences (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    notif_type      VARCHAR(100) NOT NULL,
    in_app_enabled  BOOLEAN NOT NULL DEFAULT true,
    email_enabled   BOOLEAN NOT NULL DEFAULT true,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, notif_type)
);
```

### 4.10 Tabel compliance_rules
```sql
CREATE TABLE compliance_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_type       VARCHAR(100) NOT NULL,
                    -- 'bpjs_kesehatan' | 'bpjs_tk' | 'pph21' | 'thr'
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    parameters      JSONB NOT NULL,                    -- rate, batas, dll.
    effective_from  DATE NOT NULL,
    effective_to    DATE,                              -- NULL = masih berlaku
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_compliance_rules_type ON compliance_rules(rule_type);
```

### 4.11 Tabel compliance_records
```sql
CREATE TABLE compliance_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    rule_type       VARCHAR(100) NOT NULL,
    period_year     INTEGER NOT NULL,
    period_month    INTEGER,                           -- NULL untuk tahunan
    amount_company  NUMERIC(15,2) NOT NULL DEFAULT 0,
    amount_employee NUMERIC(15,2) NOT NULL DEFAULT 0,
    status          VARCHAR(30) NOT NULL DEFAULT 'pending',
                    -- 'pending' | 'calculated' | 'paid' | 'reported'
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_compliance_records_user ON compliance_records(user_id, period_year, period_month);
```

### 4.12 Tabel fraud_logs
```sql
CREATE TABLE fraud_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id   UUID REFERENCES attendances(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    fraud_type      VARCHAR(100) NOT NULL,
                    -- 'gps_inaccurate' | 'mock_location' | 'anomaly_location' | 'anomaly_time' | 'liveness_fail'
    severity        VARCHAR(20) NOT NULL DEFAULT 'low',
                    -- 'low' | 'medium' | 'high'
    details         JSONB NOT NULL,                    -- koordinat, accuracy, metadata
    status          VARCHAR(30) NOT NULL DEFAULT 'pending',
                    -- 'pending' | 'confirmed' | 'false_positive'
    reviewed_by     UUID REFERENCES users(id),
    review_note     TEXT,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_fraud_logs_user_id ON fraud_logs(user_id);
CREATE INDEX idx_fraud_logs_status ON fraud_logs(status);
CREATE INDEX idx_fraud_logs_created_at ON fraud_logs(created_at DESC);
```

---

## 5. API Endpoints (Updated v2)

### Auth
```
POST   /api/auth/login              # Login, return JWT
POST   /api/auth/logout             # Logout (client-side token removal)
GET    /api/auth/me                 # Get current user info
```

### Employees
```
GET    /api/employees               # List semua karyawan (HR/Manager)
POST   /api/employees               # Tambah karyawan baru (HR)
GET    /api/employees/:id           # Detail karyawan
PUT    /api/employees/:id           # Update karyawan (HR)
DELETE /api/employees/:id           # Soft delete karyawan (HR)
POST   /api/employees/:id/photo     # Upload foto profil
```

### Attendance
```
POST   /api/attendance/clock-in     # Clock in (Karyawan)
POST   /api/attendance/clock-out    # Clock out (Karyawan)
GET    /api/attendance/me           # Rekap absensi sendiri
GET    /api/attendance/today        # Status absensi hari ini
GET    /api/attendance/:employee_id # Rekap absensi karyawan (HR)
PUT    /api/attendance/:id          # Override absensi (HR)
GET    /api/attendance/export/csv   # Export CSV (HR)
```

### Documents
```
POST   /api/documents/upload        # Upload dokumen baru
GET    /api/documents               # List dokumen (filter by user, kategori)
GET    /api/documents/:id           # Detail dokumen
DELETE /api/documents/:id           # Hapus dokumen (owner atau HR)
GET    /api/documents/:id/download  # Download file
POST   /api/documents/:id/comments  # Tambah komentar (HR/Manager)
GET    /api/documents/:id/comments  # List komentar
```

### AI
```
POST   /api/ai/analyze/:employee_id # Generate laporan AI
GET    /api/ai/reports/:employee_id # Riwayat laporan AI
GET    /api/ai/reports/:id          # Detail laporan AI
```

### Leave Management
```
POST   /api/leaves                  # Ajukan cuti (Karyawan)
GET    /api/leaves                  # List pengajuan cuti (filter by status, user)
GET    /api/leaves/:id              # Detail pengajuan cuti
PUT    /api/leaves/:id/approve      # Approve cuti (Manager/HR)
PUT    /api/leaves/:id/reject       # Reject cuti (Manager/HR)
PUT    /api/leaves/:id/cancel       # Cancel cuti (Karyawan, jika masih pending)
GET    /api/leaves/balance          # Saldo cuti karyawan sendiri
GET    /api/leaves/balance/:user_id # Saldo cuti karyawan (HR)
```

### Compliance
```
GET    /api/compliance/bpjs         # Kalkulasi BPJS per periode
GET    /api/compliance/pph21        # Kalkulasi PPh 21 per periode
GET    /api/compliance/thr          # Kalkulasi THR
GET    /api/compliance/checklist    # Status compliance checklist
PUT    /api/compliance/checklist/:id # Update status item checklist
```

### Notifications
```
GET    /api/notifications           # List notifikasi user (dengan pagination)
PUT    /api/notifications/:id/read  # Mark as read
PUT    /api/notifications/read-all  # Mark all as read
DELETE /api/notifications/:id       # Hapus notifikasi
GET    /api/notifications/preferences # Preference notifikasi user
PUT    /api/notifications/preferences # Update preference
GET    /api/notifications/stream    # SSE endpoint untuk real-time push
```

### Analytics & Reports
```
GET    /api/reports/attendance      # Data kehadiran untuk chart (filter: period, dept)
GET    /api/reports/summary         # Ringkasan statistik
POST   /api/reports/pdf             # Generate PDF laporan
POST   /api/reports/ai-summary      # Generate AI executive summary
```

### Fraud Detection
```
GET    /api/fraud/logs              # List fraud logs (HR)
GET    /api/fraud/logs/:id          # Detail fraud log
PUT    /api/fraud/logs/:id/review   # Update status review fraud
GET    /api/fraud/stats             # Statistik fraud per periode
```

---

## 6. SSE Architecture (Modul 7)

```
Browser                    Go Backend                  Database
  |                            |                           |
  |-- GET /api/notifications/stream (SSE) -->              |
  |                            |                           |
  |                     [Register client]                  |
  |                            |                           |
  |                     [Event loop]                       |
  |                            |                           |
  |          (event terjadi)   |                           |
  |                            |<-- INSERT notifications --|
  |                            |                           |
  |<-- data: {type, title, body} --                        |
  |                            |                           |
  |          (heartbeat 30s)   |                           |
  |<-- : ping ---------------  |                           |
```

**Implementasi Go:**
```go
// internal/notification/sse.go
type SSEHub struct {
    clients map[string]chan Notification  // key: user_id
    mu      sync.RWMutex
}

func (h *SSEHub) Subscribe(userID string) (<-chan Notification, func()) {
    ch := make(chan Notification, 10)
    h.mu.Lock()
    h.clients[userID] = ch
    h.mu.Unlock()
    
    // cleanup function
    return ch, func() {
        h.mu.Lock()
        delete(h.clients, userID)
        h.mu.Unlock()
        close(ch)
    }
}

func (h *SSEHub) Publish(userID string, n Notification) {
    h.mu.RLock()
    ch, ok := h.clients[userID]
    h.mu.RUnlock()
    if ok {
        select {
        case ch <- n:
        default: // drop jika channel penuh (client lambat)
        }
    }
}
```

---

## 7. PDF Generation (Modul 8)

Library: `github.com/jung-kurt/gofpdf` v2

```go
// internal/report/pdf.go
func GenerateAttendancePDF(data ReportData) ([]byte, error) {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    
    // Header perusahaan
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(0, 10, "Hadir — Laporan Kehadiran")
    
    // Periode
    pdf.SetFont("Arial", "", 10)
    pdf.Cell(0, 8, fmt.Sprintf("Periode: %s", data.Period))
    
    // Tabel data
    // ... render rows
    
    var buf bytes.Buffer
    err := pdf.Output(&buf)
    return buf.Bytes(), err
}
```

- **Password:** bcrypt dengan cost factor 12
- **JWT:** HS256, secret dari env, expiry 24 jam
- **File Upload:** Validasi MIME type via magic bytes, bukan hanya ekstensi
- **SQL:** Semua query menggunakan parameterized queries (pgx/v5)
- **CORS:** Hanya izinkan origin dari env `CORS_ORIGINS`
- **Rate Limiting:** Tambahkan di endpoint login (max 5 attempt/menit per IP)
- **File Path:** Sanitasi nama file, simpan di path yang tidak bisa diakses langsung via URL

---

## 7. Deployment (Development)

```bash
# Backend
cd saas-karyawan
go run cmd/server/main.go

# Frontend
cd frontend
npm run dev

# Database
docker run -d \
  --name hadir-db \
  -e POSTGRES_DB=hadir \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:16-alpine
```
