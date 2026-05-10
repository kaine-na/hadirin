# Hadir — Frontend

Frontend untuk aplikasi **Hadir** (Platform Manajemen Karyawan Digital), dibangun dengan SvelteKit 2, Svelte 5 (runes), TypeScript, dan TailwindCSS.

## Stack

- SvelteKit 2 (adapter-static, SPA mode)
- Svelte 5 dengan runes (`$state`, `$derived`, `$props`)
- TypeScript (strict mode)
- TailwindCSS 3 + `@tailwindcss/forms`
- Primary color: `#1e3a5f`

## Struktur

```
src/
├── routes/
│   ├── +layout.svelte          # Root layout (Toast container)
│   ├── +layout.ts              # SSR off, SPA mode
│   ├── +page.svelte            # Redirect ke /login atau /dashboard
│   ├── login/+page.svelte
│   └── (app)/
│       ├── +layout.svelte      # Auth guard + Sidebar + Navbar
│       ├── dashboard/+page.svelte
│       ├── attendance/+page.svelte
│       ├── attendance/manage/+page.svelte   # HR/Manager
│       ├── documents/+page.svelte
│       ├── documents/upload/+page.svelte
│       ├── documents/[id]/+page.svelte
│       ├── employees/+page.svelte           # HR/Manager
│       └── hr-ai/
│           ├── +page.svelte
│           └── [employee_id]/+page.svelte
├── lib/
│   ├── api/              # client.ts + auth/employees/attendance/documents/ai
│   ├── stores/           # auth.svelte.ts, toast.svelte.ts (runes)
│   ├── components/       # Sidebar, Navbar, Button, Badge, Table, Modal, FileUpload, Toast
│   ├── utils/format.ts   # formatDate, roleLabel, statusLabel, dsb.
│   └── types/index.ts    # TypeScript interfaces
└── app.html / app.css
```

## Menjalankan

```bash
# Install dependencies
npm install

# Dev server (port 5173, proxy /api ke backend localhost:8080)
npm run dev

# Build production (output ke ./build sebagai static SPA)
npm run build

# Preview build
npm run preview
```

## Environment

File `.env.example`:

```
PUBLIC_API_BASE_URL=http://localhost:8080
```

Saat `npm run dev`, request `/api/*` otomatis di-proxy ke `http://localhost:8080` (lihat `vite.config.ts`).

## Auth

- Token JWT disimpan di `localStorage` dengan key `sk_token`
- User info disimpan dengan key `sk_user`
- Route `/(app)/*` dilindungi via `+layout.svelte` yang memvalidasi token ke `/api/auth/me`
- 401 dari API otomatis clear auth dan redirect ke `/login`

## UI

- Bahasa Indonesia di semua label dan pesan
- Primary color palette sudah dikonfigurasi di `tailwind.config.js` (default `#1e3a5f`)
- Font: Inter
