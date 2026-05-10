# Epics & Stories — SaaS Manajemen Karyawan
**Versi:** 1.0  
**Tanggal:** 2026-05-10  
**Total:** 5 Epics, 13 Stories

---

## Epic 1: Auth & Setup
**Tujuan:** Fondasi project — Go server berjalan, database terhubung, JWT auth berfungsi, SvelteKit siap dengan routing dan auth store.

**Estimasi:** 3–4 hari

---

### Story S1.1: Setup Go Project + DB Migrations + JWT Auth
**Assignee:** Backend Dev  
**Estimasi:** 1.5 hari

**Deskripsi:**
Setup project Go dari awal dengan struktur folder yang sudah didefinisikan di architecture doc. Buat koneksi PostgreSQL, jalankan migrations, dan implementasi JWT auth endpoint.

**Tasks:**
1. Init Go module (`go mod init`)
2. Install dependencies: chi, pgx/v5, golang-jwt, bcrypt, godotenv
3. Buat `pkg/config/config.go` — load .env
4. Buat `internal/database/postgres.go` — connection pool pgx/v5
5. Buat SQL migrations (001–005)
6. Buat `internal/database/migrate.go` — auto-run migrations on startup
7. Buat `pkg/response/response.go` — standard JSON response
8. Implementasi `internal/auth/` — login, JWT issue, middleware
9. Buat `cmd/server/main.go` — setup router, start server

**Acceptance Criteria:**
- [ ] `go run cmd/server/main.go` berhasil start tanpa error
- [ ] `POST /api/auth/login` return JWT yang valid
- [ ] `GET /api/auth/me` dengan token valid return user info
- [ ] `GET /api/auth/me` tanpa token return 401
- [ ] Migrations berjalan otomatis saat server start

---

### Story S1.2: Setup SvelteKit + Routing + Auth Store
**Assignee:** Frontend Dev  
**Estimasi:** 1 hari

**Deskripsi:**
Init project SvelteKit dengan TypeScript, setup routing dasar, buat auth store dengan Svelte 5 runes, dan buat API client wrapper.

**Tasks:**
1. Init SvelteKit project di `/frontend` dengan TypeScript
2. Install dependencies: tidak ada tambahan (pakai built-in fetch)
3. Buat `lib/types/index.ts` — TypeScript interfaces (User, Attendance, Document, AIReport)
4. Buat `lib/stores/auth.ts` — Svelte 5 runes auth store
5. Buat `lib/api/client.ts` — fetch wrapper dengan auto-auth header
6. Buat `lib/api/auth.ts`, `employees.ts`, `attendance.ts`, `documents.ts`, `ai.ts`
7. Setup `+layout.svelte` — auth check, redirect ke /login jika belum login
8. Buat halaman placeholder untuk semua routes

**Acceptance Criteria:**
- [ ] `npm run dev` berhasil start di port 5173
- [ ] Akses `/dashboard` tanpa login redirect ke `/login`
- [ ] Auth store menyimpan token di localStorage
- [ ] API client otomatis tambah Authorization header

---

### Story S1.3: Login/Logout UI + Protected Routes
**Assignee:** Frontend Dev  
**Estimasi:** 0.5 hari

**Deskripsi:**
Buat halaman login yang fungsional, connect ke backend auth API, dan implementasi logout.

**Tasks:**
1. Buat `routes/login/+page.svelte` — form email + password
2. Connect form ke `lib/api/auth.ts`
3. Simpan token ke auth store setelah login berhasil
4. Redirect ke `/dashboard` setelah login
5. Buat Navbar dengan tombol logout
6. Implementasi logout — clear auth store + redirect ke /login

**Acceptance Criteria:**
- [ ] Login dengan kredensial valid berhasil masuk ke dashboard
- [ ] Login dengan kredensial salah tampilkan pesan error
- [ ] Logout berhasil clear session dan redirect ke /login
- [ ] Halaman protected tidak bisa diakses tanpa login

---

## Epic 2: Employee Management
**Tujuan:** HR bisa mengelola data karyawan — CRUD lengkap dengan UI yang intuitif.

**Estimasi:** 2–3 hari

---

### Story S2.1: Backend CRUD Employees API
**Assignee:** Backend Dev  
**Estimasi:** 1 hari

**Deskripsi:**
Implementasi semua endpoint CRUD untuk karyawan di backend Go.

**Tasks:**
1. Buat `internal/employee/model.go` — Employee struct, request/response types
2. Buat `internal/employee/repository.go` — DB queries (list, find, create, update, soft delete)
3. Buat `internal/employee/service.go` — business logic, validasi, hash password
4. Buat `internal/employee/handler.go` — HTTP handlers
5. Register routes di `main.go` dengan middleware role check (HR only untuk write)
6. Endpoint upload foto profil (simpan ke `./uploads/photos/`)

**Acceptance Criteria:**
- [ ] `GET /api/employees` return list karyawan (HR/Manager)
- [ ] `POST /api/employees` buat karyawan baru dengan password ter-hash
- [ ] `PUT /api/employees/:id` update data karyawan
- [ ] `DELETE /api/employees/:id` soft delete (is_active = false)
- [ ] Karyawan biasa tidak bisa akses endpoint ini (403)

---

### Story S2.2: Frontend Employee List + Profile Pages
**Assignee:** Frontend Dev  
**Estimasi:** 1.5 hari

**Deskripsi:**
Buat halaman daftar karyawan dengan tabel, form tambah/edit, dan halaman profil detail.

**Tasks:**
1. Buat `routes/employees/+page.svelte` — tabel karyawan dengan search/filter
2. Buat modal form tambah karyawan baru
3. Buat modal form edit karyawan
4. Implementasi konfirmasi sebelum hapus
5. Buat `routes/employees/[id]/+page.svelte` — halaman profil detail
6. Upload foto profil di halaman profil
7. Buat `lib/components/Table.svelte` — reusable table component

**Acceptance Criteria:**
- [ ] Tabel karyawan tampil dengan data dari API
- [ ] HR bisa tambah karyawan baru via form
- [ ] HR bisa edit dan hapus karyawan
- [ ] Halaman profil tampil semua info karyawan
- [ ] Karyawan biasa hanya bisa lihat profil sendiri

---

## Epic 3: Absensi
**Tujuan:** Karyawan bisa clock in/out, HR bisa monitor dan manage absensi semua karyawan.

**Estimasi:** 3–4 hari

---

### Story S3.1: Backend Attendance API
**Assignee:** Backend Dev  
**Estimasi:** 1.5 hari

**Deskripsi:**
Implementasi semua endpoint absensi — clock in/out, rekap, export CSV, dan HR override.

**Tasks:**
1. Buat `internal/attendance/model.go` — Attendance struct, request types
2. Buat `internal/attendance/repository.go` — DB queries
3. Buat `internal/attendance/service.go`:
   - Clock in: cek sudah clock in hari ini, hitung status (hadir/terlambat)
   - Clock out: cek sudah clock in, belum clock out
   - Rekap: query dengan filter tanggal, hitung summary
   - Export CSV: generate CSV string dari data rekap
4. Buat `internal/attendance/handler.go` — HTTP handlers
5. Register routes

**Acceptance Criteria:**
- [ ] Clock in dua kali dalam satu hari return error
- [ ] Status "terlambat" otomatis jika clock in > 08:00
- [ ] Export CSV menghasilkan file yang valid
- [ ] HR override tersimpan dengan field `updated_by`

---

### Story S3.2: Frontend Attendance UI
**Assignee:** Frontend Dev  
**Estimasi:** 1 hari

**Deskripsi:**
Buat halaman absensi untuk karyawan — tombol clock in/out dan tabel rekap personal.

**Tasks:**
1. Buat `routes/attendance/+page.svelte`:
   - Card status hari ini (sudah clock in/belum)
   - Tombol Clock In (disabled jika sudah clock in)
   - Tombol Clock Out (disabled jika belum clock in atau sudah clock out)
   - Tabel rekap 30 hari terakhir
2. Tampilkan waktu clock in dan clock out
3. Badge status dengan warna (Hadir=hijau, Terlambat=kuning, Izin=biru, Sakit=oranye, Alpha=merah)

**Acceptance Criteria:**
- [ ] Tombol Clock In/Out berubah state sesuai kondisi
- [ ] Rekap tampil dengan status berwarna
- [ ] Konfirmasi sebelum clock in/out

---

### Story S3.3: HR Attendance Management
**Assignee:** Frontend Dev  
**Estimasi:** 1 hari

**Deskripsi:**
Buat halaman manajemen absensi untuk HR — view semua karyawan, filter, dan override.

**Tasks:**
1. Buat `routes/attendance/manage/+page.svelte`:
   - Filter by departemen, tanggal range, status
   - Tabel semua karyawan dengan status absensi
   - Tombol export CSV
2. Modal override absensi — edit status, clock in/out time, keterangan
3. Tampilkan indikator jika record sudah di-override HR

**Acceptance Criteria:**
- [ ] HR bisa filter absensi by departemen dan tanggal
- [ ] Override tersimpan dan tampil di tabel
- [ ] Export CSV berhasil download file
- [ ] Hanya HR Admin yang bisa akses halaman ini

---

## Epic 4: Upload Berkas Kerjaan
**Tujuan:** Karyawan bisa upload dokumen kerja, HR/Manager bisa review dan beri komentar.

**Estimasi:** 3–4 hari

---

### Story S4.1: Backend Document Upload/Download API
**Assignee:** Backend Dev  
**Estimasi:** 1.5 hari

**Deskripsi:**
Implementasi endpoint upload, download, list, dan delete dokumen dengan versioning.

**Tasks:**
1. Buat `internal/document/model.go` — Document, Comment structs
2. Buat `internal/document/repository.go` — DB queries
3. Buat `internal/document/service.go`:
   - Upload: validasi MIME type, simpan file, handle versioning
   - List: query dengan filter, return latest version per title
   - Download: serve file dari filesystem
   - Delete: hapus record + file dari disk
4. Buat `internal/document/handler.go` — HTTP handlers (multipart form untuk upload)
5. Register routes

**Acceptance Criteria:**
- [ ] Upload file > 10 MB ditolak dengan error message
- [ ] File tersimpan di `./uploads/{user_id}/{uuid}_{filename}`
- [ ] Upload ulang file yang sama membuat versi baru (parent_id diset)
- [ ] Download endpoint serve file dengan Content-Type yang benar

---

### Story S4.2: Frontend Upload Form + Document List
**Assignee:** Frontend Dev  
**Estimasi:** 1 hari

**Deskripsi:**
Buat halaman daftar dokumen dan form upload.

**Tasks:**
1. Buat `routes/documents/+page.svelte` — tabel dokumen dengan filter
2. Buat `routes/documents/upload/+page.svelte` — form upload dengan drag & drop
3. Buat `lib/components/FileUpload.svelte` — reusable file upload component
4. Progress bar saat upload
5. Preview thumbnail untuk gambar

**Acceptance Criteria:**
- [ ] Form upload dengan validasi tipe file dan ukuran di client
- [ ] Progress bar tampil saat upload berlangsung
- [ ] Daftar dokumen tampil dengan info lengkap
- [ ] Filter by kategori dan tanggal berfungsi

---

### Story S4.3: HR Document Review + Komentar
**Assignee:** Frontend Dev  
**Estimasi:** 1 hari

**Deskripsi:**
Buat halaman detail dokumen dengan preview dan sistem komentar.

**Tasks:**
1. Buat `routes/documents/[id]/+page.svelte`:
   - Info dokumen (judul, kategori, tanggal, ukuran, versi)
   - Preview PDF via iframe (untuk file PDF)
   - Preview gambar (untuk file gambar)
   - Tombol download
   - Daftar versi dokumen
2. Seksi komentar — list komentar + form tambah komentar (HR/Manager)

**Acceptance Criteria:**
- [ ] PDF preview tampil di browser
- [ ] Komentar HR tersimpan dan tampil
- [ ] Daftar versi dokumen tampil dengan link ke versi lama
- [ ] Karyawan bisa lihat komentar tapi tidak bisa tambah

---

## Epic 5: AI HRD Dashboard
**Tujuan:** HR bisa generate laporan kinerja karyawan berbasis AI dengan satu klik.

**Estimasi:** 2–3 hari

---

### Story S5.1: Backend LLM Integration
**Assignee:** Backend Dev  
**Estimasi:** 1.5 hari

**Deskripsi:**
Implementasi OpenAI-compatible HTTP client dan endpoint analyze karyawan.

**Tasks:**
1. Buat `internal/ai/client.go` — HTTP client untuk OpenAI-compatible API
   - POST ke `/v1/chat/completions`
   - Handle timeout (60 detik)
   - Parse response
2. Buat `internal/ai/model.go` — AIReport struct, AnalyzeRequest
3. Buat `internal/ai/repository.go` — save dan query ai_reports
4. Buat `internal/ai/service.go`:
   - Ambil data absensi karyawan untuk periode yang diminta
   - Ambil daftar dokumen karyawan
   - Build prompt dari template
   - Call LLM client
   - Simpan hasil ke ai_reports
5. Buat `internal/ai/handler.go` — HTTP handlers
6. Register routes (hanya HR Admin)

**Acceptance Criteria:**
- [ ] `POST /api/ai/analyze/:employee_id` berhasil call LLM dan return laporan
- [ ] Laporan tersimpan ke database
- [ ] Error handling jika LLM API timeout atau error
- [ ] Hanya HR Admin yang bisa akses endpoint ini

---

### Story S5.2: Frontend AI Dashboard
**Assignee:** Frontend Dev  
**Estimasi:** 1 hari

**Deskripsi:**
Buat halaman AI HRD Dashboard — pilih karyawan, generate laporan, lihat riwayat.

**Tasks:**
1. Buat `routes/hr-ai/+page.svelte`:
   - Dropdown pilih karyawan
   - Date range picker untuk periode analisis
   - Tombol "Generate Laporan AI"
   - Loading state dengan spinner dan pesan "AI sedang menganalisis..."
2. Tampilkan hasil laporan dalam format yang mudah dibaca (markdown render)
3. Buat `routes/hr-ai/[employee_id]/+page.svelte`:
   - Riwayat laporan AI per karyawan
   - Klik laporan untuk lihat detail

**Acceptance Criteria:**
- [ ] Generate laporan berhasil dan tampil hasilnya
- [ ] Loading state tampil selama proses (bisa 5–30 detik)
- [ ] Error message tampil jika generate gagal
- [ ] Riwayat laporan tampil per karyawan
- [ ] Hanya HR Admin yang bisa akses halaman ini

---

## Summary

| Epic | Stories | Estimasi |
|------|---------|----------|
| Epic 1: Auth & Setup | 3 | 3–4 hari |
| Epic 2: Employee Management | 2 | 2–3 hari |
| Epic 3: Absensi | 3 | 3–4 hari |
| Epic 4: Upload Berkas | 3 | 3–4 hari |
| Epic 5: AI HRD | 2 | 2–3 hari |
| **Total** | **13** | **13–18 hari** |

## Urutan Implementasi yang Disarankan
1. S1.1 → S1.2 → S1.3 (fondasi dulu)
2. S2.1 → S2.2 (employee management)
3. S3.1 → S3.2 → S3.3 (absensi)
4. S4.1 → S4.2 → S4.3 (dokumen)
5. S5.1 → S5.2 (AI — terakhir karena butuh data dari modul lain)
