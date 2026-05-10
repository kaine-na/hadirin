# Epics & Stories v2 — Hadir: Platform Manajemen Karyawan Digital
**Versi:** 2.0  
**Tanggal:** 2026-05-10  
**Status:** Ready for Sprint Planning

---

## Ringkasan Estimasi Total

| Epic | Modul | Stories | Estimasi Total |
|------|-------|---------|----------------|
| Epic 5 | Smart Leave Management | 5 stories | ~13 hari |
| Epic 6 | Compliance Engine Indonesia | 5 stories | ~15 hari |
| Epic 7 | Smart Notification System | 4 stories | ~10 hari |
| Epic 8 | HR Analytics & Reporting | 5 stories | ~14 hari |
| Epic 9 | Deteksi Fraud Absensi | 5 stories | ~16 hari |
| **Total** | | **24 stories** | **~68 hari kerja** |

Estimasi effort: S = 1-2 hari, M = 3-5 hari, L = 6-10 hari

---

## Epic 5 — Smart Leave Management

**Goal:** Karyawan bisa mengajukan cuti secara digital dengan approval workflow multi-level, saldo cuti otomatis, dan integrasi ke rekap absensi.

**Definition of Done:**
- Karyawan bisa submit, lihat status, dan cancel pengajuan cuti
- Manager dan HR bisa approve/reject dengan catatan
- Saldo cuti terupdate otomatis
- Hari cuti muncul di rekap absensi sebagai status "Cuti"

---

### Story 5.1 — Database & Model Leave Management
**Effort:** M (3 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai developer, saya perlu membuat schema database dan model Go untuk modul leave management.

**Tasks:**
- Buat migration `006_create_leave_requests.sql` (tabel leave_requests)
- Buat migration `007_create_leave_balances.sql` (tabel leave_balances)
- Buat `internal/leave/model.go` dengan struct LeaveRequest, LeaveBalance, LeaveType
- Buat `internal/leave/repository.go` dengan CRUD queries

**Acceptance Criteria:**
- [ ] Migration berjalan tanpa error
- [ ] Constraint UNIQUE(user_id, year, leave_type) di leave_balances berfungsi
- [ ] Saldo awal 12 hari/tahun otomatis dibuat saat karyawan baru ditambahkan
- [ ] Pro-rata dihitung berdasarkan bulan bergabung (joined_at)

---

### Story 5.2 — API Leave Request (Submit & Cancel)
**Effort:** M (4 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai Karyawan, saya bisa mengajukan cuti baru dan membatalkan pengajuan yang masih pending.

**Endpoints:**
- `POST /api/leaves` — submit pengajuan cuti
- `PUT /api/leaves/:id/cancel` — cancel (hanya jika status pending)
- `GET /api/leaves` — list pengajuan cuti milik sendiri
- `GET /api/leaves/balance` — saldo cuti sendiri

**Acceptance Criteria:**
- [ ] Validasi: tanggal mulai tidak boleh di masa lalu
- [ ] Validasi: total hari tidak melebihi saldo tersisa
- [ ] Status awal selalu "pending"
- [ ] Cancel hanya bisa jika status masih "pending"
- [ ] Saldo tidak berkurang saat submit, hanya saat approved_hr

---

### Story 5.3 — API Leave Approval (Manager & HR)
**Effort:** M (3 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai Manager dan HR Admin, saya bisa menyetujui atau menolak pengajuan cuti dengan catatan.

**Endpoints:**
- `PUT /api/leaves/:id/approve` — approve (Manager → approved_manager, HR → approved_hr)
- `PUT /api/leaves/:id/reject` — reject dengan catatan
- `GET /api/leaves` — list semua pengajuan (HR/Manager, dengan filter status)

**Acceptance Criteria:**
- [ ] Manager hanya bisa approve ke status "approved_manager"
- [ ] HR hanya bisa approve ke "approved_hr" jika sudah "approved_manager"
- [ ] Saat HR approve, saldo cuti berkurang otomatis
- [ ] Saat HR approve, record absensi dibuat untuk setiap hari cuti dengan status "cuti"
- [ ] Reject bisa dilakukan di state manapun oleh Manager atau HR

---

### Story 5.4 — Frontend Leave Request Page
**Effort:** L (6 hari)  
**Assignee:** Frontend

**Deskripsi:**  
Sebagai Karyawan, saya bisa melihat saldo cuti, mengajukan cuti baru, dan melihat riwayat pengajuan.

**Route:** `/leaves`

**UI Components:**
- Card saldo cuti per jenis (tahunan, sakit, dll.)
- Form pengajuan cuti (jenis, tanggal mulai-selesai, alasan)
- Tabel riwayat pengajuan dengan status badge
- Tombol cancel untuk pengajuan pending

**Acceptance Criteria:**
- [ ] Saldo cuti tampil akurat per jenis
- [ ] Form validasi client-side (tanggal, saldo)
- [ ] Status badge: pending (kuning), approved_manager (biru), approved_hr (hijau), rejected (merah), cancelled (abu)
- [ ] Tombol cancel hanya muncul untuk status pending

---

### Story 5.5 — Frontend Leave Management (HR/Manager)
**Effort:** M (3 hari)  
**Assignee:** Frontend

**Deskripsi:**  
Sebagai HR Admin dan Manager, saya bisa melihat semua pengajuan cuti dan melakukan approval/rejection.

**Route:** `/leaves/manage`

**UI Components:**
- Tabel semua pengajuan dengan filter status dan karyawan
- Modal approve dengan field catatan
- Modal reject dengan field alasan wajib
- Badge status yang jelas

**Acceptance Criteria:**
- [ ] Filter by status dan karyawan berfungsi
- [ ] Approve/reject memperbarui status secara real-time
- [ ] Catatan HR/Manager tersimpan dan tampil di detail
- [ ] Halaman hanya accessible oleh HR Admin dan Manager

---

## Epic 6 — Compliance Engine Indonesia

**Goal:** HR bisa menghitung BPJS, PPh 21, dan THR secara otomatis sesuai regulasi Indonesia terbaru, dengan compliance checklist dan notifikasi deadline.

**Definition of Done:**
- Kalkulasi BPJS akurat sesuai regulasi
- PPh 21 menggunakan metode TER (PMK 168/2023)
- THR dihitung pro-rata berdasarkan masa kerja
- Compliance checklist dengan status real-time

---

### Story 6.1 — Database & Model Compliance
**Effort:** S (2 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai developer, saya perlu membuat schema database untuk compliance rules dan records.

**Tasks:**
- Buat migration `008_create_compliance_rules.sql`
- Buat migration `009_create_compliance_records.sql`
- Seed data: aturan BPJS, PPh 21 TER, THR sesuai regulasi 2024
- Buat `internal/compliance/model.go`

**Acceptance Criteria:**
- [ ] Tabel compliance_rules berisi data regulasi terbaru
- [ ] Tabel compliance_records siap menyimpan kalkulasi per karyawan per periode
- [ ] Seed data bisa di-update tanpa mengubah kode

---

### Story 6.2 — Kalkulasi BPJS
**Effort:** M (4 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai HR Admin, saya bisa melihat kalkulasi BPJS Kesehatan dan Ketenagakerjaan per karyawan per bulan.

**Endpoints:**
- `GET /api/compliance/bpjs?year=2026&month=5` — kalkulasi BPJS semua karyawan
- `GET /api/compliance/bpjs/:user_id?year=2026&month=5` — kalkulasi per karyawan

**Acceptance Criteria:**
- [ ] BPJS Kesehatan: 4% perusahaan + 1% karyawan, batas Rp 12jt
- [ ] BPJS TK JHT: 3.7% + 2%, JP: 2% + 1%, JKK: 0.24%, JKM: 0.3%
- [ ] Kalkulasi berdasarkan gaji pokok di profil karyawan
- [ ] Response berisi breakdown per komponen

---

### Story 6.3 — Kalkulasi PPh 21 (Metode TER)
**Effort:** L (6 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai HR Admin, saya bisa melihat kalkulasi PPh 21 menggunakan metode TER sesuai PMK 168/2023.

**Endpoints:**
- `GET /api/compliance/pph21?year=2026&month=5` — kalkulasi PPh 21 semua karyawan

**Acceptance Criteria:**
- [ ] Kategori TER A/B/C berdasarkan status pernikahan dan tanggungan
- [ ] Tarif TER sesuai tabel PMK 168/2023
- [ ] Rekonsiliasi tahunan di bulan Desember
- [ ] Hasil kalkulasi tersimpan di compliance_records

---

### Story 6.4 — Kalkulasi THR
**Effort:** M (3 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai HR Admin, saya bisa generate laporan THR untuk semua karyawan menjelang hari raya.

**Endpoints:**
- `GET /api/compliance/thr?holiday=lebaran&year=2026` — kalkulasi THR

**Acceptance Criteria:**
- [ ] Karyawan > 12 bulan: 1x gaji pokok
- [ ] Karyawan < 12 bulan: pro-rata (bulan kerja / 12 x gaji pokok)
- [ ] Masa kerja dihitung dari joined_at
- [ ] Export ke CSV tersedia

---

### Story 6.5 — Frontend Compliance Dashboard
**Effort:** L (6 hari)  
**Assignee:** Frontend

**Deskripsi:**  
Sebagai HR Admin, saya bisa melihat dashboard compliance dengan checklist status dan kalkulasi per periode.

**Route:** `/compliance`

**UI Components:**
- Compliance checklist dengan status hijau/kuning/merah
- Tab: BPJS | PPh 21 | THR
- Tabel kalkulasi per karyawan
- Tombol export CSV per komponen

**Acceptance Criteria:**
- [ ] Checklist menampilkan status real-time per item
- [ ] Tabel kalkulasi bisa difilter per periode
- [ ] Export CSV berfungsi untuk setiap komponen
- [ ] Halaman hanya accessible oleh HR Admin

---

## Epic 7 — Smart Notification System

**Goal:** User menerima notifikasi real-time in-app via SSE dan email via SMTP, dengan preference setting per user.

**Definition of Done:**
- SSE terhubung dan push notifikasi real-time
- Email terkirim untuk event penting
- Notification center dengan read/unread state
- User bisa atur preference notifikasi

---

### Story 7.1 — Database & SSE Hub Backend
**Effort:** M (4 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai developer, saya perlu membuat infrastructure notifikasi: database schema, SSE hub, dan service layer.

**Tasks:**
- Buat migration `010_create_notifications.sql`
- Buat migration `011_create_notification_preferences.sql`
- Implementasi `internal/notification/sse.go` (SSEHub dengan subscribe/publish)
- Implementasi `internal/notification/service.go` (create, list, mark read)

**Acceptance Criteria:**
- [ ] SSEHub thread-safe dengan sync.RWMutex
- [ ] Channel per user dengan buffer size 10
- [ ] Cleanup otomatis saat client disconnect
- [ ] Heartbeat ping setiap 30 detik untuk keep-alive

---

### Story 7.2 — API Notification Endpoints
**Effort:** M (3 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai User, saya bisa mengakses notifikasi via REST API dan SSE stream.

**Endpoints:**
- `GET /api/notifications/stream` — SSE endpoint
- `GET /api/notifications` — list notifikasi (pagination)
- `PUT /api/notifications/:id/read` — mark as read
- `PUT /api/notifications/read-all` — mark all as read
- `DELETE /api/notifications/:id` — hapus notifikasi
- `GET /api/notifications/preferences` — get preference
- `PUT /api/notifications/preferences` — update preference

**Acceptance Criteria:**
- [ ] SSE endpoint mengirim event dengan format `data: {json}\n\n`
- [ ] List notifikasi ter-paginate dengan filter is_read
- [ ] Preference tersimpan per user per tipe notifikasi

---

### Story 7.3 — Event Triggers Integration
**Effort:** M (4 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai developer, saya perlu mengintegrasikan notifikasi ke semua event yang relevan di sistem.

**Tasks:**
- Inject NotificationService ke handler leave, document, attendance
- Trigger notifikasi saat: leave submitted, leave approved/rejected, document commented
- Buat cron job untuk clock_in_reminder (07:45 WIB setiap hari kerja)
- Buat cron job untuk compliance_deadline (H-7 dan H-1)

**Acceptance Criteria:**
- [ ] Notifikasi terkirim dalam < 1 detik setelah event
- [ ] Cron job berjalan tepat waktu
- [ ] Notifikasi tidak duplikat jika event dipanggil berkali-kali

---

### Story 7.4 — Frontend Notification Center
**Effort:** M (3 hari)  
**Assignee:** Frontend

**Deskripsi:**  
Sebagai User, saya bisa melihat notifikasi real-time di notification center dengan badge counter.

**UI Components:**
- Bell icon di Navbar dengan badge counter (unread count)
- Dropdown notification center (max 10 terbaru)
- Link ke halaman `/notifications` untuk semua notifikasi
- Halaman `/notifications` dengan list lengkap + preference setting

**Acceptance Criteria:**
- [ ] SSE terhubung otomatis saat login
- [ ] Badge counter terupdate real-time tanpa refresh
- [ ] Klik notifikasi → mark as read + navigate ke halaman terkait
- [ ] Preference toggle tersimpan ke backend

---

## Epic 8 — HR Analytics & Reporting

**Goal:** HR bisa melihat dashboard analytics kehadiran dengan chart, filter, export PDF, dan AI executive summary.

**Definition of Done:**
- Dashboard analytics dengan chart real-time
- Filter periode dan departemen berfungsi
- Export PDF dengan layout profesional
- AI summary relevan dengan data

---

### Story 8.1 — Analytics API Endpoints
**Effort:** M (4 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai developer, saya perlu membuat API endpoints untuk data analytics kehadiran.

**Endpoints:**
- `GET /api/reports/attendance?period=month&dept=engineering` — data kehadiran untuk chart
- `GET /api/reports/summary` — ringkasan statistik (total karyawan, rata-rata kehadiran, dll.)
- `GET /api/reports/late-trend?period=month` — tren keterlambatan per minggu

**Acceptance Criteria:**
- [ ] Response berisi data yang siap di-render sebagai chart
- [ ] Filter periode: week, month, quarter, custom (start_date, end_date)
- [ ] Filter departemen berfungsi
- [ ] Query dioptimasi dengan index yang tepat

---

### Story 8.2 — PDF Generation Service
**Effort:** L (6 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai HR Admin, saya bisa generate laporan kehadiran dalam format PDF dengan header perusahaan.

**Endpoint:**
- `POST /api/reports/pdf` — generate PDF, return file

**Tasks:**
- Install `github.com/jung-kurt/gofpdf`
- Implementasi `internal/report/pdf.go`
- Template: header perusahaan, tabel data, footer

**Acceptance Criteria:**
- [ ] PDF ter-generate dalam < 10 detik untuk data 1 bulan
- [ ] Header berisi nama perusahaan "Hadir" dan periode laporan
- [ ] Tabel berisi: nama, departemen, total hadir, terlambat, izin, sakit, alpha
- [ ] Footer berisi tanggal generate dan nama HR

---

### Story 8.3 — AI Executive Summary
**Effort:** M (3 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai HR Admin, saya bisa generate AI executive summary dari data laporan kehadiran.

**Endpoint:**
- `POST /api/reports/ai-summary` — generate AI summary

**Acceptance Criteria:**
- [ ] AI menganalisis data dan menghasilkan narasi dalam Bahasa Indonesia
- [ ] Highlight anomali (departemen dengan kehadiran rendah, tren memburuk)
- [ ] Rekomendasi tindakan yang konkret
- [ ] Summary tersimpan ke database untuk referensi

---

### Story 8.4 — Frontend Analytics Dashboard
**Effort:** L (7 hari)  
**Assignee:** Frontend

**Deskripsi:**  
Sebagai HR Admin, saya bisa melihat dashboard analytics dengan chart interaktif dan filter.

**Route:** `/reports`

**UI Components:**
- Chart kehadiran per departemen (bar chart — gunakan Chart.js atau Recharts)
- Tren keterlambatan (line chart)
- Distribusi status absensi (pie chart)
- Filter periode dan departemen
- Tombol export PDF dan generate AI summary

**Acceptance Criteria:**
- [ ] Chart ter-render dengan data real dari API
- [ ] Filter mengubah data chart tanpa reload halaman
- [ ] Loading state saat fetch data
- [ ] Halaman hanya accessible oleh HR Admin dan Manager

---

### Story 8.5 — PDF Download & AI Summary UI
**Effort:** S (2 hari)  
**Assignee:** Frontend

**Deskripsi:**  
Sebagai HR Admin, saya bisa download PDF laporan dan melihat AI executive summary di halaman reports.

**UI Components:**
- Tombol "Export PDF" → trigger download
- Tombol "Generate AI Summary" → loading state → tampilkan hasil
- Card AI summary dengan format yang rapi

**Acceptance Criteria:**
- [ ] PDF ter-download otomatis saat tombol diklik
- [ ] AI summary tampil dalam card dengan format markdown
- [ ] Loading state selama generate (bisa 5-30 detik)
- [ ] Error handling jika AI API tidak tersedia

---

## Epic 9 — Deteksi Fraud Absensi

**Goal:** Sistem mendeteksi kecurangan absensi secara otomatis via GPS validation, mock location detection, dan anomaly detection, dengan fraud dashboard untuk HR.

**Definition of Done:**
- Clock-in dengan GPS tidak akurat ditolak
- Anomaly detection berjalan otomatis
- Fraud logs tersimpan dengan detail
- HR bisa review dan update status fraud

---

### Story 9.1 — Database & Model Fraud Detection
**Effort:** S (2 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai developer, saya perlu membuat schema database untuk fraud logs dan update model attendance.

**Tasks:**
- Buat migration `012_create_fraud_logs.sql`
- Update migration attendances: tambah kolom latitude, longitude, gps_accuracy, selfie_url
- Buat `internal/fraud/model.go` dan `repository.go`

**Acceptance Criteria:**
- [ ] Tabel fraud_logs dengan semua kolom yang diperlukan
- [ ] Tabel attendances memiliki kolom GPS dan selfie
- [ ] Index pada fraud_logs(user_id) dan fraud_logs(status)

---

### Story 9.2 — GPS Validation & Clock-in Enhancement
**Effort:** M (4 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai Sistem, clock-in dengan GPS accuracy > 100m otomatis ditolak dan dicatat sebagai fraud attempt.

**Tasks:**
- Update `POST /api/attendance/clock-in` untuk menerima latitude, longitude, gps_accuracy
- Validasi: reject jika accuracy > 100m
- Simpan koordinat GPS di tabel attendances
- Buat fraud log jika GPS tidak valid

**Acceptance Criteria:**
- [ ] Request tanpa GPS data masih diterima (backward compatible)
- [ ] GPS accuracy > 100m → HTTP 422 dengan pesan error yang jelas
- [ ] Koordinat GPS tersimpan untuk audit trail
- [ ] Fraud log dibuat dengan type "gps_inaccurate" dan detail accuracy

---

### Story 9.3 — Anomaly Detection Engine
**Effort:** L (7 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai Sistem, pola absensi anomali otomatis di-flag saat clock-in.

**Anomali yang dideteksi:**
1. Clock-in dari 2 lokasi berbeda dalam 5 menit (distance > 1km)
2. Clock-in di luar jam kerja normal (sebelum 06:00 atau setelah 22:00)
3. Koordinat GPS di laut atau di luar bounding box Indonesia
4. Koordinat tidak berubah sama sekali selama 7 hari berturut-turut

**Acceptance Criteria:**
- [ ] Setiap anomali menghasilkan fraud log dengan severity yang tepat
- [ ] Anomali tidak memblokir clock-in (hanya flag, bukan reject)
- [ ] Bounding box Indonesia: lat -11 s/d 6, lon 95 s/d 141
- [ ] Deteksi berjalan async (tidak memperlambat response clock-in)

---

### Story 9.4 — Fraud API Endpoints
**Effort:** M (3 hari)  
**Assignee:** Backend

**Deskripsi:**  
Sebagai HR Admin, saya bisa mengakses fraud logs via API dan mengupdate status review.

**Endpoints:**
- `GET /api/fraud/logs` — list fraud logs (filter: user, date, type, status)
- `GET /api/fraud/logs/:id` — detail fraud log
- `PUT /api/fraud/logs/:id/review` — update status (confirmed/false_positive) + catatan
- `GET /api/fraud/stats` — statistik fraud per periode

**Acceptance Criteria:**
- [ ] List fraud logs ter-paginate
- [ ] Filter by severity, type, status, date range berfungsi
- [ ] Review update menyimpan reviewer_id dan reviewed_at
- [ ] Stats berisi: total fraud, per type, per severity

---

### Story 9.5 — Frontend Fraud Dashboard
**Effort:** L (6 hari)  
**Assignee:** Frontend

**Deskripsi:**  
Sebagai HR Admin, saya bisa melihat fraud dashboard dengan semua fraud attempts dan melakukan review.

**Route:** `/fraud`

**UI Components:**
- Stats cards: total fraud, pending review, confirmed, false positive
- Tabel fraud logs dengan filter dan pagination
- Modal review: update status + tambah catatan
- Badge severity: low (kuning), medium (oranye), high (merah)

**Acceptance Criteria:**
- [ ] Stats cards terupdate real-time
- [ ] Filter by severity, type, status, karyawan berfungsi
- [ ] Modal review memperbarui status tanpa reload halaman
- [ ] Halaman hanya accessible oleh HR Admin

---

## Prioritas Sprint Rekomendasi

### Sprint 1 (2 minggu) — Foundation
- Story 5.1 (Leave DB)
- Story 7.1 (Notification DB + SSE Hub)
- Story 9.1 (Fraud DB)
- Story 8.1 (Analytics API)

### Sprint 2 (2 minggu) — Core Features
- Story 5.2 (Leave Submit API)
- Story 5.3 (Leave Approval API)
- Story 7.2 (Notification API)
- Story 9.2 (GPS Validation)

### Sprint 3 (2 minggu) — Frontend Wave 1
- Story 5.4 (Leave Frontend)
- Story 5.5 (Leave Management Frontend)
- Story 7.4 (Notification Center)
- Story 8.4 (Analytics Dashboard)

### Sprint 4 (2 minggu) — Compliance + Fraud
- Story 6.1-6.4 (Compliance Backend)
- Story 9.3 (Anomaly Detection)
- Story 9.4 (Fraud API)

### Sprint 5 (2 minggu) — Polish & Advanced
- Story 6.5 (Compliance Frontend)
- Story 7.3 (Event Triggers)
- Story 8.2-8.3 (PDF + AI Summary)
- Story 9.5 (Fraud Dashboard)
- Story 8.5 (PDF Download UI)
