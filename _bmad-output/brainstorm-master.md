# Master Brainstorming — Hadir
Date: 2026-05-10

---

## Nama Produk

### Rekomendasi Utama: HADIR

**Mengapa Hadir menang:**

Hadir adalah satu-satunya nama yang mendapat rekomendasi kuat dari 2 agent sekaligus (Nova dan Scout), dan secara konsisten muncul di top 3 semua analisis. Keunggulan utamanya:

- **Double meaning yang sempurna:** "Hadir" = absensi (fitur inti produk) + "kami hadir untuk Anda" (brand promise layanan). Tidak ada nama lain yang punya dua makna relevan sekaligus.
- **Satu kata, sangat memorable:** Lebih kuat dari KaryaHub (dua kata), lebih relevan dari Sigap, lebih mudah diingat dari Karyaku.
- **Positioning hangat dan human:** Berbeda dari kompetitor yang terasa korporat atau teknis. Cocok untuk UKM yang butuh HR tools yang approachable.
- **Developer-friendly:** Forge menilai "Hadirku" (varian) 9/10 — pendek, fonetik jelas, tidak ada konflik package. Domain `hadir.id` atau `hadirku.id` kemungkinan tersedia.
- **Tagline natural:** "Hadir untuk tim kamu, setiap hari." — langsung komunikatif.

**Skor perbandingan:**

| Kriteria | Hadir | Karyaku | Sigap |
|----------|-------|---------|-------|
| Branding strength | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| Kemudahan diingat | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| Relevansi produk | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| Potensi domain | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| Dukungan agent | Nova ✓, Scout ✓ | Scout ✓ | Nova ✓ |

### Alternatif:

1. **Karyaku** — "Karya" (hasil kerja) + "-ku" (milikku). Employee-centric positioning yang unik, SEO potential tinggi. Direkomendasikan Scout sebagai top pick. Cocok jika ingin positioning lebih ke arah karyawan daripada HR admin.

2. **Sigap** — Satu kata berenergi, berarti "cepat, tanggap, siap siaga". Direkomendasikan Nova. Cocok jika ingin positioning yang lebih dinamis dan tech-forward. Tagline: "HR yang sigap, bisnis yang bergerak cepat."

---

## Top 5 Fitur Baru (Prioritas Implementasi)

Dipilih berdasarkan Impact vs Effort matrix, relevansi pasar Indonesia, dan feasibility teknis (stack SvelteKit + Go + PostgreSQL).

---

### 1. Smart Leave Management — Impact: HIGH | Effort: M (5-8 hari)

**Sumber:** Forge (fitur teknis) + Memo (Tier 1 wajib ada)

**Mengapa prioritas #1:**
Manajemen cuti adalah fitur WAJIB yang belum ada di produk saat ini. Ini bukan fitur baru — ini gap yang harus ditutup sebelum produk bisa dianggap HRIS lengkap. Semua kompetitor punya ini. Tanpa cuti, produk tidak bisa dijual ke perusahaan yang butuh compliance UU Ketenagakerjaan.

**Deskripsi:**
Sistem pengajuan dan approval cuti terintegrasi dengan data absensi. Karyawan ajukan cuti (tahunan, sakit, izin khusus), manager/HR approve/reject, sistem otomatis update status absensi. AI bisa rekomendasikan apakah cuti layak disetujui berdasarkan pola historis.

**Implementasi teknis (Forge):**
- Tabel `leave_requests` dengan state machine (pending → approved/rejected)
- 7 endpoint baru: POST/GET /api/leaves, PUT approve/reject, GET balance
- Komponen: LeaveRequestForm, LeaveApprovalCard, LeaveBalanceWidget
- Effort: M (3-5 hari backend + 2-3 hari frontend)

---

### 2. Compliance Engine Indonesia — Impact: HIGH | Effort: M (7-10 hari)

**Sumber:** Scout (Autopilot Regulasi) + Memo (Compliance Engine)

**Mengapa prioritas #2:**
Pain point terbesar UKM Indonesia. Regulasi HR sangat kompleks (PPh 21 TER, BPJS, THR, UMR per provinsi) dan sering berubah. Tidak ada kompetitor SME yang punya "autopilot" yang proaktif update ketika regulasi berubah. Ini adalah differentiator unik yang langsung terasa ROI-nya — perusahaan rela bayar premium untuk menghilangkan risiko denda.

**Deskripsi:**
Modul otomasi kepatuhan regulasi: kalkulasi BPJS otomatis, PPh 21 dengan semua metode, THR multi-agama, pesangon/UPMK sesuai UU Cipta Kerja. Notifikasi proaktif ketika ada perubahan regulasi. Generate file siap upload ke portal BPJS, DJP Online, Coretax.

**Implementasi teknis:**
- Rule-based engine (tidak butuh ML)
- Database regulasi yang bisa di-update berkala
- Integrasi e-Filing DJP untuk laporan PPh 21
- Compliance checklist per periode dengan status hijau/kuning/merah

---

### 3. Smart Notification System — Impact: HIGH | Effort: M (5-6 hari)

**Sumber:** Forge (Notifikasi & Reminder Otomatis) + Memo (Tier 2 high demand)

**Mengapa prioritas #3:**
Fitur yang meningkatkan engagement dan retention pengguna secara signifikan. Reminder clock-in, notifikasi approval, alert karyawan terlambat — semua ini mengurangi pekerjaan manual HR. Effort relatif kecil tapi impact besar pada daily active usage produk.

**Deskripsi:**
Sistem notifikasi in-app + email untuk event penting: reminder clock-in (jam 08:00 jika belum absen), notifikasi approval dokumen/cuti, reminder jatah cuti yang akan habis, alert karyawan sering terlambat. Real-time via SSE.

**Implementasi teknis (Forge):**
- Tabel `notifications` di PostgreSQL
- SSE endpoint untuk real-time push ke frontend
- Cron goroutine di background (time.Ticker)
- Email via gomail
- Komponen: NotificationBell, NotificationDropdown di navbar
- Effort: M (3-4 hari backend + 2 hari frontend)

---

### 4. HR Analytics & Reporting (Export PDF) — Impact: MEDIUM | Effort: L (7-10 hari)

**Sumber:** Forge (HR Analytics & Reporting) + Memo (Dashboard HR analytics real-time)

**Mengapa prioritas #4:**
Dashboard analytics yang lebih kaya adalah fitur yang langsung terlihat oleh decision maker (owner/CEO) saat demo. Chart interaktif dan export PDF membuat produk terasa "enterprise" tanpa kompleksitas enterprise. AI executive summary dari data laporan adalah differentiator yang memanfaatkan AI yang sudah ada.

**Deskripsi:**
Dashboard analytics dengan chart interaktif (kehadiran per departemen, tren keterlambatan, distribusi status) dan export laporan ke PDF. AI generate executive summary dari data laporan. Filter periode, departemen, status.

**Implementasi teknis (Forge):**
- Endpoint agregasi baru untuk data chart
- PDF generation: go-pdf/fpdf atau gofpdf
- Frontend: Chart.js + svelte-chartjs wrapper (~60KB)
- Komponen: AttendanceChart, DepartmentPieChart, ReportFilter, ExportButton
- Route baru: /reports
- Effort: L (4-6 hari backend + 3-4 hari frontend)

---

### 5. Deteksi Fraud Absensi + Validasi Lokasi Cerdas — Impact: HIGH | Effort: M (5-7 hari)

**Sumber:** Scout (Deteksi Fraud Absensi) + Memo (Anomaly Detection)

**Mengapa prioritas #5:**
GPS spoofing untuk absensi sudah jadi masalah umum di Indonesia, terutama untuk karyawan lapangan. Ini adalah pain point nyata yang belum diselesaikan secara komprehensif di segmen UKM. Fitur ini langsung menjawab kekhawatiran owner/manajer tentang kecurangan absensi — sangat powerful untuk closing sales.

**Deskripsi:**
Deteksi fake GPS otomatis (GPS accuracy check, mock location detection, device fingerprinting). Validasi multi-layer: GPS + WiFi network recognition + selfie dengan liveness detection. Anomaly detection berbasis AI: flagging pola absensi mencurigakan. Dashboard fraud report untuk HR.

**Implementasi teknis:**
- Memanfaatkan AI yang sudah ada untuk anomaly detection
- Liveness detection untuk selfie (anti-foto)
- Offline attendance mode dengan sync otomatis
- Dashboard fraud report dengan bukti foto dan koordinat

---

## Ringkasan Kontribusi Per Agent

### Nova (Marketing & Branding)
Menghasilkan 10 kandidat nama dengan analisis branding mendalam. Top 3: Hadir (double meaning kuat), KaryaHub (lokal + modern), Sigap (berenergi). Fokus pada resonansi emosional, kemudahan diingat, dan diferensiasi dari kompetitor. Rekomendasi utama: **Hadir**.

### Scout (Market Research)
Menganalisis 14 produk kompetitor (7 lokal + 7 global), mengidentifikasi 8 tren naming HR SaaS, dan menemukan gap pasar nyata. Menghasilkan 5 rekomendasi nama dan 5 fitur berdasarkan gap kompetitor. Key insight: HRIS untuk UKM Indonesia yang sedang tumbuh — tidak terlalu sederhana seperti Kerjoo, tidak terlalu kompleks seperti Talenta. Rekomendasi nama: **Karyaku**, runner-up **Hadir**.

### Memo (Research & Knowledge)
Riset mendalam tren HR Tech 2025-2026 global dan Indonesia. Mengidentifikasi 10 tren dominan, pain point UKM Indonesia, dan 8 teknologi AI/ML relevan. Menghasilkan 7 ide fitur inovatif dengan nilai bisnis dan kompleksitas implementasi. Rekomendasi nama dari bahasa Jawa/Indonesia kuno: **Karsa** (kehendak untuk bertindak).

### Forge (Technical)
Review codebase aktual dan mengidentifikasi 3 quick wins (rate limiting, validasi AI period, pagination), 2 tech debt kritis (JWT blacklist, file storage migration), dan 5 ide fitur baru dengan spesifikasi teknis lengkap (endpoint, komponen, estimasi effort). Rekomendasi nama dari perspektif developer: **Hadirku** (9/10) dan **Rekap.id** (9/10).

---

## Quick Wins yang Harus Dikerjakan Sekarang

Sebelum fitur baru, Forge mengidentifikasi 3 quick wins yang bisa dikerjakan segera:

1. **Rate Limiting Login** — Effort: S (<1 jam), Impact: HIGH (security). Endpoint login rentan brute force.
2. **Validasi period_start <= period_end di AI** — Effort: S (<30 menit), Impact: MEDIUM (UX).
3. **Proper Pagination Frontend** — Effort: S-M (2-4 jam), Impact: MEDIUM (correctness).

Dan 2 tech debt kritis sebelum launch/scale:
- **JWT Token Blacklist** — Effort: M (1-2 hari), wajib sebelum production
- **File Storage Migration ke Object Storage** — Effort: L (3-5 hari), wajib sebelum scale

---

## Keputusan yang Dibutuhkan dari User

1. **Nama produk:** Hadir (rekomendasi) vs Karyaku vs Sigap?
2. **Fitur pertama yang diimplementasikan:** Smart Leave Management (paling mendesak) atau fitur lain?
3. **Quick wins:** Langsung kerjakan 3 quick wins sekarang sebelum fitur baru?
