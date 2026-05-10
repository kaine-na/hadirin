# Project Context — SaaS Karyawan

## Identitas Project
- **Nama:** SaaS Manajemen Karyawan
- **Tipe:** Web Application (Multi-tenant SaaS)
- **Bahasa UI:** Indonesia

## Tech Stack
| Layer | Teknologi |
|-------|-----------|
| Frontend | SvelteKit (Svelte 5) |
| Backend | Go (Golang) |
| Database | PostgreSQL |
| ORM/Query | pgx/v5 (native PostgreSQL driver) |
| Auth | JWT (JSON Web Token) |
| File Storage | Local filesystem (`./uploads/`) |
| LLM | Custom OpenAI-compatible API |

## LLM Configuration
- **Base URL:** `https://api.openai.com/v1` (configurable via `LLM_BASE_URL`)
- **API Key:** set via `LLM_API_KEY` environment variable
- **Model:** `claude-sonnet-4.6`
- **Penggunaan:** AI HRD Dashboard — analisis kinerja karyawan

## Fitur Utama
1. **Auth & User Management** — Login JWT, role-based access, CRUD karyawan
2. **Absensi** — Clock in/out, rekap, export CSV, HR override
3. **Upload Berkas Kerjaan** — Upload dokumen, versioning, preview, komentar HR
4. **AI HRD Dashboard** — LLM generate laporan kinerja karyawan

## Role & Akses
| Role | Akses |
|------|-------|
| Super Admin | Full access, manage semua tenant |
| HR Admin | Manage karyawan, absensi, dokumen, AI dashboard |
| Manager | View karyawan di departemennya, review dokumen |
| Karyawan | Clock in/out, upload berkas, lihat rekap sendiri |

## Konvensi Kode
- Backend: Go idioms, error handling eksplisit, no panic di production
- Frontend: Svelte 5 runes, TypeScript, komponen kecil dan fokus
- API: RESTful, JSON response standar `{success, data, message}`
- Database: SQL migrations manual di `/migrations/`
- Env config: `.env` file, tidak pernah hardcode secrets

## Struktur Direktori
```
saas-karyawan/
├── cmd/server/main.go
├── internal/
│   ├── auth/
│   ├── employee/
│   ├── attendance/
│   ├── document/
│   ├── ai/
│   └── database/
├── pkg/
│   ├── config/
│   └── response/
├── migrations/
├── uploads/
├── frontend/
│   └── src/
│       ├── routes/
│       └── lib/
├── _bmad-output/
└── .env.example
```

## Constraints & Keputusan Arsitektur
- File storage lokal (bukan S3/cloud) untuk simplisitas awal
- JWT stateless — tidak ada session store
- Multi-tenant via `company_id` di setiap tabel utama
- LLM call bersifat async — response disimpan ke `ai_reports`
- Export CSV dihasilkan on-the-fly, tidak disimpan
