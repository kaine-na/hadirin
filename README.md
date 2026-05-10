# Hadir — Sistem Manajemen Karyawan Berbasis AI

> **Hadir** adalah platform HRIS (Human Resource Information System) berbasis web untuk perusahaan skala kecil-menengah Indonesia. Dibangun dengan SvelteKit + Go, dilengkapi AI assistant untuk insight HR, dan fitur compliance regulasi Indonesia secara otomatis.

![Hadir Banner](https://img.shields.io/badge/Hadir-HRIS%20Indonesia-blue?style=for-the-badge)
![SvelteKit](https://img.shields.io/badge/SvelteKit-5-FF3E00?style=flat-square&logo=svelte)
![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat-square&logo=postgresql)
![TailwindCSS](https://img.shields.io/badge/TailwindCSS-3-38B2AC?style=flat-square&logo=tailwind-css)

---

## ✨ Fitur Utama

### 👤 Manajemen Karyawan & Auth
- Login berbasis JWT (access token 24 jam + refresh token 7 hari)
- Role: Super Admin, HR Admin, Manager, Karyawan
- CRUD karyawan: nama, NIK, jabatan, departemen, foto profil
- Soft delete, status aktif/non-aktif

### 🕐 Absensi Digital
- Clock In / Clock Out dengan timestamp otomatis
- Deteksi keterlambatan (batas jam 08:00)
- Rekap absensi harian, mingguan, bulanan
- Export rekap ke CSV
- Override/koreksi absensi oleh HR

### 📁 Manajemen Dokumen Kerja
- Upload dokumen laporan kerja (PDF, DOCX, XLSX)
- Sistem komentar per dokumen
- Approval workflow dokumen
- Riwayat versi dokumen

### 🤖 AI HRD Dashboard
- Analisis kinerja karyawan berbasis AI
- Insight otomatis dari data absensi dan dokumen
- Rekomendasi tindakan untuk HR
- Executive summary laporan

### 🏖️ Manajemen Cuti (Smart Leave)
- Pengajuan cuti: tahunan, sakit, izin khusus, melahirkan
- Approval workflow: Karyawan → Manager → HR
- State machine: pending → approved / rejected / cancelled
- Saldo cuti otomatis (12 hari/tahun, pro-rata)
- AI rekomendasi approval berdasarkan pola historis tim

### 🔔 Notifikasi Real-time
- Notifikasi in-app via SSE (Server-Sent Events)
- Email notification via SMTP
- Reminder clock-in otomatis (jam 08:15)
- Alert approval dokumen & cuti
- Background worker untuk reminder berkala

### 📊 HR Analytics & Laporan
- Dashboard chart interaktif (Chart.js)
- Tren kehadiran per departemen
- Distribusi status absensi (pie chart)
- Top karyawan terlambat
- Export laporan ke PDF dengan header perusahaan
- AI executive summary dari data laporan

### ⚖️ Compliance Engine Indonesia
- Kalkulasi BPJS Kesehatan (4% perusahaan + 1% karyawan)
- Kalkulasi BPJS Ketenagakerjaan (JHT, JP, JKK, JKM)
- PPh 21 metode TER sesuai PMK 168/2023
- Kalkulasi THR otomatis (pro-rata untuk masa kerja < 12 bulan)
- Compliance checklist bulanan dengan status hijau/kuning/merah
- Notifikasi proaktif H-3 sebelum deadline pelaporan

### 🛡️ Deteksi Fraud Absensi
- Validasi akurasi GPS (tolak jika > 100 meter)
- Deteksi mock location / GPS spoofing
- Selfie saat clock-in dengan liveness check
- Velocity check: deteksi 2 lokasi berbeda dalam waktu singkat
- Anomaly detection berbasis AI
- Dashboard fraud report untuk HR

---

## 🏗️ Arsitektur

```
hadir/
├── frontend/          # SvelteKit 5 + TailwindCSS 3
│   ├── src/
│   │   ├── routes/    # Halaman aplikasi
│   │   │   ├── (app)/
│   │   │   │   ├── dashboard/
│   │   │   │   ├── attendance/
│   │   │   │   ├── documents/
│   │   │   │   ├── employees/
│   │   │   │   ├── leaves/
│   │   │   │   ├── reports/
│   │   │   │   ├── compliance/
│   │   │   │   ├── fraud/
│   │   │   │   └── hr-ai/
│   │   │   └── login/
│   │   ├── lib/
│   │   │   ├── components/  # 20+ komponen reusable
│   │   │   ├── api/         # API modules per domain
│   │   │   ├── stores/      # Svelte 5 runes stores
│   │   │   └── utils/       # Helper functions
│   │   └── app.html
│   └── package.json
│
├── backend/           # Go 1.22 + Gin + GORM
│   ├── cmd/server/    # Entry point
│   ├── internal/
│   │   ├── auth/          # JWT auth + rate limiting
│   │   ├── employee/      # CRUD karyawan
│   │   ├── attendance/    # Absensi + fraud detection
│   │   ├── document/      # Manajemen dokumen
│   │   ├── leave/         # Manajemen cuti
│   │   ├── notification/  # SSE + background worker
│   │   ├── analytics/     # HR analytics + PDF
│   │   ├── compliance/    # BPJS + PPh21 + THR
│   │   ├── fraud/         # GPS + liveness + anomaly
│   │   ├── ai/            # AI integration
│   │   ├── middleware/    # Auth + CORS + rate limit
│   │   └── router/        # Route definitions
│   ├── migrations/    # 10 SQL migrations
│   └── go.mod
│
├── docker-compose.yml
└── _bmad-output/      # Dokumentasi BMAD (PRD, Architecture, Epics)
```

---

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- Node.js 18+
- PostgreSQL 14+ (atau gunakan embedded PostgreSQL via Docker)

### 1. Clone Repository

```bash
git clone https://github.com/kaine-na/hadirin.git
cd hadirin
```

### 2. Setup Backend

```bash
cd backend

# Copy environment variables
cp .env.example .env

# Edit .env sesuai konfigurasi lokal
nano .env

# Download dependencies
go mod download

# Jalankan server (migrations otomatis berjalan)
go run cmd/server/main.go
```

Backend berjalan di `http://localhost:8081`

### 3. Setup Frontend

```bash
cd frontend

# Install dependencies
npm install

# Copy environment variables
cp .env.example .env

# Jalankan dev server
npm run dev
```

Frontend berjalan di `http://localhost:5173`

### 4. Menggunakan Docker Compose

```bash
# Jalankan semua service sekaligus
docker-compose up -d

# Lihat logs
docker-compose logs -f
```

---

## ⚙️ Konfigurasi

### Backend `.env`

```env
# Server
APP_NAME=Hadir
PORT=8081
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=hadir

# JWT
JWT_SECRET=your-super-secret-key-here
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h

# AI Provider (OpenAI-compatible)
AI_BASE_URL=https://api.openai.com/v1
AI_API_KEY=your-api-key
AI_MODEL=gpt-4o

# Storage
UPLOAD_DIR=./uploads
MAX_FILE_SIZE=10485760  # 10MB

# SMTP (opsional, untuk email notifikasi)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@email.com
SMTP_PASS=your-app-password
```

### Frontend `.env`

```env
PUBLIC_API_URL=http://localhost:8081
```

---

## 📡 API Endpoints

### Auth
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/auth/login` | Login dengan email + password |
| POST | `/api/auth/logout` | Logout (invalidasi token) |
| POST | `/api/auth/refresh` | Refresh access token |
| GET | `/api/auth/me` | Data user yang sedang login |

### Karyawan
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/employees` | List karyawan (paginated) |
| POST | `/api/employees` | Tambah karyawan baru |
| GET | `/api/employees/:id` | Detail karyawan |
| PUT | `/api/employees/:id` | Update data karyawan |
| DELETE | `/api/employees/:id` | Soft delete karyawan |
| POST | `/api/employees/:id/photo` | Upload foto profil |

### Absensi
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/attendance/clock-in` | Clock in (+ selfie + GPS) |
| POST | `/api/attendance/clock-out` | Clock out |
| GET | `/api/attendance` | Rekap absensi |
| GET | `/api/attendance/today` | Status absensi hari ini |
| PUT | `/api/attendance/:id` | Koreksi absensi (HR only) |
| GET | `/api/attendance/export` | Export CSV |

### Cuti
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/leaves` | Ajukan cuti |
| GET | `/api/leaves` | List pengajuan cuti |
| GET | `/api/leaves/balance` | Saldo cuti user |
| PUT | `/api/leaves/:id/approve` | Approve cuti (HR/Manager) |
| PUT | `/api/leaves/:id/reject` | Reject cuti |
| PUT | `/api/leaves/:id/cancel` | Cancel cuti |
| GET | `/api/leaves/:id/ai-recommendation` | Rekomendasi AI |

### Notifikasi
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/notifications/stream` | SSE stream real-time |
| GET | `/api/notifications` | List notifikasi |
| GET | `/api/notifications/unread-count` | Jumlah unread |
| PUT | `/api/notifications/read-all` | Tandai semua dibaca |
| PUT | `/api/notifications/:id/read` | Tandai satu dibaca |

### Analytics & Laporan
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/analytics/attendance-summary` | Ringkasan kehadiran |
| GET | `/api/analytics/department-stats` | Statistik per departemen |
| GET | `/api/analytics/trend` | Tren 30 hari |
| GET | `/api/analytics/top-late-employees` | Top karyawan terlambat |
| GET | `/api/reports/export-pdf` | Export laporan PDF |

### Compliance
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/compliance/bpjs-calculation` | Kalkulasi BPJS |
| GET | `/api/compliance/pph21-calculation` | Kalkulasi PPh 21 TER |
| GET | `/api/compliance/thr-calculation` | Kalkulasi THR |
| GET | `/api/compliance/checklist` | Checklist kepatuhan |
| PUT | `/api/compliance/checklist/:id/done` | Tandai checklist selesai |

### Fraud Detection
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/fraud/logs` | Log fraud (HR only) |
| GET | `/api/fraud/summary` | Ringkasan fraud bulan ini |
| PUT | `/api/fraud/logs/:id/dismiss` | Dismiss false positive |
| PUT | `/api/fraud/logs/:id/confirm` | Konfirmasi fraud |

---

## 🗄️ Database Schema

### Tabel Utama
- `users` — data karyawan + auth
- `attendances` — rekap absensi harian
- `documents` — dokumen kerja
- `document_comments` — komentar dokumen
- `ai_reports` — hasil analisis AI
- `leave_types` — jenis cuti
- `leave_balances` — saldo cuti per karyawan
- `leave_requests` — pengajuan cuti
- `notifications` — notifikasi in-app
- `compliance_records` — hasil kalkulasi compliance
- `fraud_logs` — log deteksi fraud
- `attendance_photos` — foto selfie clock-in

---

## 🔐 Role & Permission

| Fitur | Karyawan | Manager | HR Admin | Super Admin |
|-------|----------|---------|----------|-------------|
| Lihat profil sendiri | ✅ | ✅ | ✅ | ✅ |
| Edit profil sendiri | ✅ | ✅ | ✅ | ✅ |
| Clock in/out | ✅ | ✅ | ✅ | ✅ |
| Lihat absensi sendiri | ✅ | ✅ | ✅ | ✅ |
| Lihat absensi semua | ❌ | ✅ | ✅ | ✅ |
| Koreksi absensi | ❌ | ❌ | ✅ | ✅ |
| Ajukan cuti | ✅ | ✅ | ✅ | ✅ |
| Approve cuti | ❌ | ✅ | ✅ | ✅ |
| CRUD karyawan | ❌ | ❌ | ✅ | ✅ |
| Lihat analytics | ❌ | ✅ | ✅ | ✅ |
| Compliance engine | ❌ | ❌ | ✅ | ✅ |
| Fraud dashboard | ❌ | ❌ | ✅ | ✅ |
| AI HRD Dashboard | ❌ | ✅ | ✅ | ✅ |

---

## 🧪 Default Credentials (Development)

```
Email: admin@company.com
Password: admin123
Role: HR Admin
```

> ⚠️ Ganti credentials ini sebelum deploy ke production!

---

## 🛠️ Tech Stack

### Frontend
- **SvelteKit 5** — framework dengan Svelte 5 runes
- **TailwindCSS 3** — utility-first CSS
- **Lucide Svelte** — icon library
- **Chart.js** — visualisasi data interaktif
- **TypeScript** — type safety

### Backend
- **Go 1.22** — bahasa pemrograman utama
- **Gin** — HTTP web framework
- **GORM** — ORM untuk PostgreSQL
- **JWT** — autentikasi stateless
- **SSE** — Server-Sent Events untuk notifikasi real-time
- **gofpdf** — generate laporan PDF

### Database
- **PostgreSQL 14+** — database utama
- **10 migrations** — schema versioning

### Infrastructure
- **Docker + Docker Compose** — containerization
- **Nginx** — reverse proxy (production)
- **Cloudflare Tunnel** — expose ke internet tanpa port forwarding

---

## 📋 Development Workflow

Project ini dibangun menggunakan **BMAD Method** (Business-driven Multi-Agent Development):

1. **PRD** — Product Requirements Document (`_bmad-output/prd.md`)
2. **Architecture** — Desain arsitektur sistem (`_bmad-output/architecture.md`)
3. **Epics & Stories** — Breakdown implementasi (`_bmad-output/epics-v2.md`)
4. **Implementation** — Coding per modul
5. **Code Review** — Review otomatis + security scan (`CODE_REVIEW.md`)

---

## 📄 Lisensi

MIT License — bebas digunakan untuk keperluan komersial maupun personal.

---

## 👨‍💻 Author

**Nadi Rifqi** — AI-Assisted Software Engineer  
Portfolio: [naddrzz.pages.dev](https://naddrzz.pages.dev)  
GitHub: [@kaine-na](https://github.com/kaine-na)
