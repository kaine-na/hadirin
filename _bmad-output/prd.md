# PRD — Hadir: Platform Manajemen Karyawan Digital
**Versi:** 2.0  
**Tanggal:** 2026-05-10  
**Status:** Draft — siap implementasi

---

## 1. Overview

SaaS Manajemen Karyawan adalah aplikasi web berbasis multi-tenant yang membantu perusahaan mengelola kehadiran karyawan, dokumen kerja, dan mendapatkan insight kinerja berbasis AI. Target pengguna adalah perusahaan skala kecil-menengah (10–500 karyawan).

### Problem Statement
- HR masih mengelola absensi manual (spreadsheet/kertas)
- Dokumen laporan kerja tersebar, sulit dilacak
- Tidak ada insight otomatis tentang kinerja karyawan

### Solution
Aplikasi terpusat dengan fitur absensi digital, manajemen dokumen, dan AI assistant untuk HR.

---

## 2. Modul 1 — Auth & User Management

### 2.1 Deskripsi
Sistem autentikasi berbasis JWT dengan manajemen role dan profil karyawan.

### 2.2 User Stories
- Sebagai HR Admin, saya bisa login dengan email dan password
- Sebagai HR Admin, saya bisa menambah, edit, dan hapus data karyawan
- Sebagai Karyawan, saya bisa melihat dan edit profil saya sendiri
- Sebagai Super Admin, saya bisa mengelola semua user di sistem

### 2.3 Fitur Detail

**Login & Auth**
- Form login dengan email + password
- JWT access token (expire: 24 jam) + refresh token (expire: 7 hari)
- Logout invalidasi token di client
- Proteksi route berdasarkan role

**Manajemen User**
- CRUD karyawan: nama, email, password, jabatan, departemen, foto
- Role: Super Admin, HR Admin, Manager, Karyawan
- Status karyawan: Aktif / Non-aktif
- Soft delete (tidak hapus permanen dari DB)

**Profil Karyawan**
- Foto profil (upload gambar)
- Informasi: nama lengkap, NIK, jabatan, departemen, tanggal bergabung
- HR bisa edit semua field; Karyawan hanya bisa edit foto dan info kontak

### 2.4 Acceptance Criteria
- [ ] Login berhasil menghasilkan JWT yang valid
- [ ] Route yang dilindungi menolak request tanpa token
- [ ] HR Admin bisa CRUD karyawan
- [ ] Karyawan tidak bisa akses data karyawan lain

---

## 3. Modul 2 — Absensi

### 3.1 Deskripsi
Sistem clock in/out digital dengan rekap dan kemampuan export.

### 3.2 User Stories
- Sebagai Karyawan, saya bisa clock in saat mulai kerja
- Sebagai Karyawan, saya bisa clock out saat selesai kerja
- Sebagai Karyawan, saya bisa melihat rekap absensi saya
- Sebagai HR Admin, saya bisa melihat absensi semua karyawan
- Sebagai HR Admin, saya bisa override/koreksi data absensi
- Sebagai HR Admin, saya bisa export rekap ke CSV

### 3.3 Fitur Detail

**Clock In / Clock Out**
- Tombol Clock In hanya aktif jika belum clock in hari ini
- Tombol Clock Out hanya aktif setelah clock in
- Timestamp otomatis saat tombol ditekan
- Catat IP address dan user agent device
- Batas waktu clock in: jam 08:00 — lewat dari itu status "Terlambat"

**Status Absensi**
- Hadir: clock in sebelum batas waktu
- Terlambat: clock in setelah batas waktu
- Izin: diinput manual oleh HR dengan keterangan
- Sakit: diinput manual oleh HR dengan keterangan
- Alpha: tidak ada record clock in (otomatis di akhir hari)

**Rekap Absensi**
- View harian: daftar karyawan + status hari ini
- View mingguan: tabel 7 hari per karyawan
- View bulanan: ringkasan per karyawan (total hadir, terlambat, izin, sakit, alpha)
- Filter by departemen, tanggal range

**Export CSV**
- Export rekap bulanan per karyawan atau semua karyawan
- Format: Nama, NIK, Departemen, Tanggal, Clock In, Clock Out, Status, Keterangan

**HR Override**
- HR bisa edit status, clock in time, clock out time, dan keterangan
- Log perubahan tersimpan (siapa yang ubah, kapan, nilai lama vs baru)

### 3.4 Acceptance Criteria
- [ ] Karyawan tidak bisa clock in dua kali dalam satu hari
- [ ] Status "Terlambat" otomatis jika clock in > 08:00
- [ ] Export CSV menghasilkan file yang bisa dibuka di Excel
- [ ] HR override tersimpan dengan audit log

---

## 4. Modul 3 — Upload Berkas Kerjaan

### 4.1 Deskripsi
Sistem manajemen dokumen untuk laporan kerja karyawan dengan versioning dan review HR.

### 4.2 User Stories
- Sebagai Karyawan, saya bisa upload file laporan kerja
- Sebagai Karyawan, saya bisa melihat daftar dokumen yang sudah saya upload
- Sebagai HR/Manager, saya bisa melihat dokumen semua karyawan
- Sebagai HR/Manager, saya bisa memberi komentar pada dokumen
- Sebagai Karyawan, saya bisa upload versi baru dari dokumen yang sama

### 4.3 Fitur Detail

**Upload Dokumen**
- Format yang didukung: PDF, DOCX, XLSX, PNG, JPG, JPEG
- Ukuran maksimal: 10 MB per file
- Metadata wajib: judul, deskripsi, kategori, tanggal dokumen
- Kategori: Laporan Harian, Laporan Mingguan, Laporan Proyek, Lainnya

**Daftar Dokumen**
- List dokumen dengan filter: karyawan, kategori, tanggal range
- Tampilkan: judul, kategori, tanggal upload, ukuran file, versi
- Sort by: tanggal terbaru, nama, kategori

**Versioning**
- Upload ulang file dengan judul yang sama = versi baru
- Semua versi tersimpan dan bisa diakses
- Tampilkan versi terbaru secara default, ada opsi lihat versi lama

**Preview & Download**
- PDF: preview langsung di browser (iframe/embed)
- Gambar: preview thumbnail
- DOCX/XLSX: hanya download (tidak ada preview)

**Komentar HR/Manager**
- HR/Manager bisa tambah komentar teks pada dokumen
- Komentar tampil di halaman detail dokumen
- Karyawan bisa baca komentar tapi tidak bisa hapus

### 4.4 Acceptance Criteria
- [ ] File tersimpan di `./uploads/{user_id}/{filename}`
- [ ] Upload file > 10 MB ditolak dengan pesan error
- [ ] Versioning: upload ulang tidak menghapus versi lama
- [ ] Komentar tersimpan dan tampil di detail dokumen

---

## 5. Modul 4 — AI HRD Dashboard

### 5.1 Deskripsi
Dashboard berbasis AI untuk HR menganalisis kinerja karyawan menggunakan LLM.

### 5.2 User Stories
- Sebagai HR Admin, saya bisa memilih karyawan dan generate laporan kinerja AI
- Sebagai HR Admin, saya bisa melihat riwayat laporan AI yang sudah dibuat
- Sebagai HR Admin, saya bisa melihat ringkasan kinerja, pola kehadiran, dan rekomendasi

### 5.3 Fitur Detail

**Generate Laporan AI**
- HR pilih karyawan dari dropdown
- Pilih periode analisis (bulan/range tanggal)
- Klik "Generate Laporan" — sistem kirim data ke LLM
- Loading state selama proses (bisa 5–30 detik)
- Hasil tampil di halaman setelah selesai

**Input ke LLM**
- Data absensi: total hadir, terlambat, izin, sakit, alpha dalam periode
- Daftar dokumen: judul, kategori, tanggal, jumlah dokumen per kategori
- Profil karyawan: nama, jabatan, departemen, lama bekerja

**Output dari LLM**
- Ringkasan kinerja kehadiran
- Analisis produktivitas berdasarkan dokumen yang diupload
- Pola kehadiran (konsisten/tidak, hari-hari bermasalah)
- Rekomendasi HR (tindakan yang disarankan)

**Riwayat Laporan**
- Semua laporan AI tersimpan di database
- HR bisa lihat laporan lama per karyawan
- Tampilkan: tanggal generate, periode analisis, nama HR yang generate

**Prompt Template**
```
Kamu adalah asisten HR profesional. Analisis data karyawan berikut dan berikan laporan kinerja dalam Bahasa Indonesia.

Karyawan: {nama} | Jabatan: {jabatan} | Departemen: {departemen}
Periode: {periode}

DATA ABSENSI:
- Total hari kerja: {total_hari}
- Hadir: {hadir} hari
- Terlambat: {terlambat} hari  
- Izin: {izin} hari
- Sakit: {sakit} hari
- Alpha: {alpha} hari

DOKUMEN YANG DIUPLOAD:
{daftar_dokumen}

Berikan analisis mencakup:
1. Ringkasan kinerja kehadiran
2. Analisis produktivitas dari dokumen
3. Pola yang perlu diperhatikan
4. Rekomendasi untuk HR
```

### 5.4 Acceptance Criteria
- [ ] Generate laporan berhasil memanggil LLM API
- [ ] Laporan tersimpan ke database setelah generate
- [ ] Riwayat laporan tampil per karyawan
- [ ] Error handling jika LLM API tidak tersedia

---

## 6. Non-Functional Requirements

### Performance
- Halaman load < 2 detik (tanpa AI call)
- Upload file < 5 detik untuk file 10 MB
- AI generate laporan: timeout 60 detik

### Security
- Password di-hash dengan bcrypt (cost factor 12)
- JWT secret dari environment variable
- File upload: validasi MIME type, bukan hanya ekstensi
- SQL injection prevention via parameterized queries
- CORS dikonfigurasi hanya untuk domain yang diizinkan

### Reliability
- Error handling di semua endpoint API
- Graceful degradation jika LLM API down (tampilkan pesan error, bukan crash)
- Database connection pooling

---

## 7. Out of Scope (v1)
- Mobile app (hanya web)
- Notifikasi email/push
- Geolocation untuk absensi
- Integrasi payroll
- Multi-bahasa (hanya Indonesia)
- Real-time updates (WebSocket)

---

## 8. Modul 5 — Smart Leave Management

### 8.1 Deskripsi
Sistem manajemen cuti digital dengan approval workflow multi-level dan saldo cuti otomatis.

### 8.2 User Stories
- Sebagai Karyawan, saya bisa mengajukan cuti dengan memilih jenis dan tanggal
- Sebagai Manager, saya bisa menyetujui atau menolak pengajuan cuti karyawan di tim saya
- Sebagai HR Admin, saya bisa melihat semua pengajuan cuti dan melakukan approval final
- Sebagai Karyawan, saya bisa melihat saldo cuti saya yang tersisa
- Sebagai HR Admin, saya bisa melihat kalender cuti tim untuk perencanaan

### 8.3 Fitur Detail

**Jenis Cuti**
- Cuti Tahunan: 12 hari/tahun, pro-rata untuk karyawan baru
- Cuti Sakit: tidak terbatas, butuh surat dokter untuk > 2 hari
- Izin Khusus: pernikahan (3 hari), duka cita (2 hari), dll.
- Cuti Melahirkan: 3 bulan untuk ibu, 2 hari untuk ayah

**Approval Workflow**
- State machine: `pending` → `approved_manager` → `approved_hr` / `rejected` / `cancelled`
- Karyawan submit → notifikasi ke Manager
- Manager approve → notifikasi ke HR untuk approval final
- HR approve/reject → notifikasi ke Karyawan
- Karyawan bisa cancel selama masih `pending`

**Saldo Cuti**
- Saldo tahunan otomatis di-reset setiap 1 Januari
- Pro-rata untuk karyawan baru (berdasarkan bulan bergabung)
- Saldo berkurang otomatis saat cuti diapprove
- Saldo dikembalikan jika cuti di-cancel setelah approve

**Integrasi Absensi**
- Hari cuti yang diapprove otomatis tercatat sebagai status "Cuti" di rekap absensi
- Tidak dihitung sebagai alpha atau terlambat

**AI Rekomendasi**
- Berdasarkan pola historis tim, AI merekomendasikan apakah cuti sebaiknya diapprove
- Faktor: beban kerja tim, musim sibuk, riwayat cuti karyawan

### 8.4 Acceptance Criteria
- [ ] Karyawan tidak bisa mengajukan cuti melebihi saldo yang tersisa
- [ ] Approval workflow berjalan sesuai state machine
- [ ] Hari cuti yang diapprove muncul di rekap absensi sebagai "Cuti"
- [ ] Saldo cuti terupdate otomatis setelah approval

---

## 9. Modul 6 — Compliance Engine Indonesia

### 9.1 Deskripsi
Mesin kalkulasi kepatuhan regulasi Indonesia: BPJS, PPh 21, dan THR otomatis.

### 9.2 User Stories
- Sebagai HR Admin, saya bisa melihat kalkulasi BPJS per karyawan per bulan
- Sebagai HR Admin, saya bisa melihat kalkulasi PPh 21 dengan metode TER
- Sebagai HR Admin, saya bisa generate laporan THR menjelang Lebaran/Natal
- Sebagai HR Admin, saya mendapat notifikasi deadline pelaporan pajak dan BPJS

### 9.3 Fitur Detail

**BPJS Kesehatan**
- Iuran perusahaan: 4% dari gaji pokok
- Iuran karyawan: 1% dari gaji pokok
- Batas atas gaji untuk kalkulasi: Rp 12.000.000/bulan
- Laporan bulanan per karyawan

**BPJS Ketenagakerjaan**
- JHT: 3.7% perusahaan + 2% karyawan
- JP: 2% perusahaan + 1% karyawan
- JKK: 0.24% perusahaan (standar)
- JKM: 0.3% perusahaan

**PPh 21 (Metode TER — PMK 168/2023)**
- Tarif Efektif Rata-rata berdasarkan penghasilan bruto bulanan
- Kategori TER: A (tidak kawin), B (kawin tanpa tanggungan), C (kawin + tanggungan)
- Kalkulasi otomatis berdasarkan data profil karyawan
- Rekonsiliasi tahunan di bulan Desember

**THR (Tunjangan Hari Raya)**
- Karyawan > 12 bulan: 1x gaji pokok
- Karyawan < 12 bulan: pro-rata (bulan kerja / 12 x gaji pokok)
- Deadline: 7 hari sebelum hari raya
- Laporan THR per karyawan

**Compliance Checklist**
- Status per periode: hijau (selesai), kuning (mendekati deadline), merah (terlambat)
- Item: laporan BPJS bulanan, SPT PPh 21, pembayaran THR
- Notifikasi proaktif H-7 dan H-1 sebelum deadline

### 9.4 Acceptance Criteria
- [ ] Kalkulasi BPJS akurat sesuai regulasi terbaru
- [ ] PPh 21 menggunakan metode TER sesuai PMK 168/2023
- [ ] THR pro-rata dihitung berdasarkan tanggal bergabung
- [ ] Compliance checklist menampilkan status real-time

---

## 10. Modul 7 — Smart Notification System

### 10.1 Deskripsi
Sistem notifikasi real-time in-app via SSE dan email via SMTP dengan preference per user.

### 10.2 User Stories
- Sebagai Karyawan, saya menerima reminder clock-in setiap pagi
- Sebagai Manager, saya mendapat notifikasi saat ada pengajuan cuti baru
- Sebagai Karyawan, saya mendapat notifikasi saat dokumen saya dikomentari
- Sebagai User, saya bisa mengatur notifikasi mana yang ingin saya terima

### 10.3 Fitur Detail

**Notifikasi In-App (SSE)**
- Server-Sent Events untuk push real-time ke browser
- Notification center dengan badge counter
- Status read/unread per notifikasi
- Hapus notifikasi individual atau semua

**Email Notification (SMTP)**
- Template email HTML untuk setiap jenis event
- Konfigurasi SMTP via environment variable
- Retry otomatis jika pengiriman gagal (max 3x)

**Event Triggers**
- `clock_in_reminder`: setiap hari kerja jam 07:45 WIB
- `leave_request_submitted`: ke Manager saat ada pengajuan cuti
- `leave_approved`: ke Karyawan saat cuti diapprove
- `leave_rejected`: ke Karyawan saat cuti ditolak
- `document_commented`: ke pemilik dokumen saat ada komentar baru
- `attendance_late_alert`: ke HR saat karyawan terlambat > 30 menit
- `compliance_deadline`: ke HR H-7 dan H-1 sebelum deadline

**Preference Setting**
- Toggle per tipe notifikasi (in-app dan/atau email)
- Jam quiet hours (tidak kirim notifikasi di luar jam kerja)
- Simpan preference di database per user

### 10.4 Acceptance Criteria
- [ ] SSE terhubung dan menerima notifikasi real-time tanpa refresh
- [ ] Email terkirim dalam < 30 detik setelah event
- [ ] Preference user tersimpan dan dihormati
- [ ] Notification center menampilkan badge count yang akurat

---

## 11. Modul 8 — HR Analytics & Reporting

### 11.1 Deskripsi
Dashboard analytics kehadiran dan produktivitas dengan export PDF dan AI executive summary.

### 11.2 User Stories
- Sebagai HR Admin, saya bisa melihat chart kehadiran per departemen
- Sebagai HR Admin, saya bisa filter data berdasarkan periode dan departemen
- Sebagai HR Admin, saya bisa export laporan ke PDF dengan header perusahaan
- Sebagai HR Admin, saya bisa mendapat AI executive summary dari data laporan

### 11.3 Fitur Detail

**Dashboard Analytics**
- Chart kehadiran per departemen (bar chart)
- Tren keterlambatan per minggu/bulan (line chart)
- Distribusi status absensi (pie chart)
- Top 5 karyawan paling sering terlambat
- Ringkasan: total karyawan aktif, rata-rata kehadiran, total cuti bulan ini

**Filter & Drill-down**
- Filter periode: minggu ini, bulan ini, kuartal ini, custom range
- Filter departemen: semua atau per departemen
- Filter status karyawan: aktif/non-aktif
- Drill-down: klik departemen → lihat detail per karyawan

**Export PDF**
- Header perusahaan (nama, logo placeholder, periode laporan)
- Tabel data kehadiran per karyawan
- Chart yang di-render sebagai gambar
- Footer dengan tanggal generate dan nama HR

**AI Executive Summary**
- Tombol "Generate AI Summary" di halaman laporan
- AI menganalisis data dan menghasilkan narasi ringkasan
- Highlight anomali dan rekomendasi tindakan
- Simpan summary ke database untuk referensi

**Route Baru**
- `/reports` — halaman utama analytics
- `/reports/attendance` — detail kehadiran
- `/reports/compliance` — status compliance

### 11.4 Acceptance Criteria
- [ ] Chart ter-render dengan data real dari database
- [ ] Filter mengubah data chart secara real-time
- [ ] PDF ter-generate dengan layout yang rapi
- [ ] AI summary relevan dengan data yang ditampilkan

---

## 12. Modul 9 — Deteksi Fraud Absensi

### 12.1 Deskripsi
Sistem deteksi kecurangan absensi berbasis GPS, metadata device, dan anomaly detection.

### 12.2 User Stories
- Sebagai HR Admin, saya bisa melihat laporan absensi yang terdeteksi mencurigakan
- Sebagai Sistem, absensi dengan GPS tidak akurat otomatis ditolak
- Sebagai Sistem, pola absensi anomali otomatis di-flag untuk review HR
- Sebagai HR Admin, saya bisa melihat detail fraud attempt per karyawan

### 12.3 Fitur Detail

**GPS Accuracy Validation**
- Saat clock-in, frontend kirim koordinat GPS + accuracy radius
- Backend reject jika accuracy > 100 meter
- Simpan koordinat untuk audit trail
- Tampilkan peta lokasi di detail absensi (HR view)

**Mock Location Detection**
- Deteksi via device metadata (user agent analysis)
- Flag jika koordinat tidak berubah sama sekali selama beberapa hari
- Deteksi VPN/proxy via IP geolocation mismatch

**Selfie Liveness Check**
- Karyawan wajib selfie saat clock-in
- Liveness check sederhana: deteksi blink (kedip mata)
- Simpan foto sebagai bukti audit
- HR bisa review foto per absensi

**Anomaly Detection**
- Clock-in dari 2 lokasi berbeda dalam 5 menit → flag otomatis
- Clock-in di luar jam kerja normal (sebelum 06:00 atau setelah 22:00) → flag
- Pola clock-in yang terlalu konsisten (detik yang sama setiap hari) → flag
- Koordinat GPS di laut atau di luar Indonesia → reject

**Fraud Report Dashboard**
- Daftar semua fraud attempt dengan severity (low/medium/high)
- Filter by karyawan, tanggal, jenis fraud
- Status: pending review / confirmed fraud / false positive
- HR bisa tambah catatan per fraud attempt

### 12.4 Acceptance Criteria
- [ ] Clock-in dengan GPS accuracy > 100m ditolak dengan pesan error
- [ ] Anomaly detection berjalan otomatis saat clock-in
- [ ] Fraud dashboard menampilkan semua flag dengan detail
- [ ] HR bisa update status fraud attempt

---

## 13. Non-Functional Requirements (Updated v2)

### Performance
- Halaman load < 2 detik (tanpa AI call)
- Upload file < 5 detik untuk file 10 MB
- AI generate laporan: timeout 60 detik
- SSE connection: reconnect otomatis jika terputus
- PDF generation: < 10 detik untuk laporan 1 bulan

### Security
- Password di-hash dengan bcrypt (cost factor 12)
- JWT secret dari environment variable
- File upload: validasi MIME type, bukan hanya ekstensi
- SQL injection prevention via parameterized queries
- CORS dikonfigurasi hanya untuk domain yang diizinkan
- Rate limiting: max 5 request/menit per IP untuk endpoint login
- GPS data: validasi server-side, tidak hanya client-side

### Reliability
- Error handling di semua endpoint API
- Graceful degradation jika LLM API down
- Database connection pooling
- SSE reconnection dengan exponential backoff

---

## 14. Out of Scope (v2)
- Mobile app native (iOS/Android)
- Multi-tenant (satu instance per perusahaan)
- Integrasi payroll eksternal (Gadjian, Talenta)
- Biometric fingerprint hardware
- Video call untuk interview
