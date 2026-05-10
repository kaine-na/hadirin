Read _bmad-output/prd.md and _bmad-output/architecture.md first.

You are redesigning the SaaS Karyawan frontend. The project is at:
/home/nadi/.hermes/workspace/shared/projects/saas-karyawan/frontend

## Your Mission
Complete frontend redesign with:
1. Install lucide-svelte: `cd frontend && npm install lucide-svelte`
2. Redesign ALL pages with modern, professional UI (anti-generic)
3. Add Lucide icons throughout (sidebar, buttons, status badges, page headers)
4. Add interactive effects: hover lift on cards, smooth transitions (150ms), button press scale, page fade-in animation
5. Add visual feedback: skeleton loaders, toast notifications, empty states with CTA, form validation, confirmation modals
6. Add micro-interactions: Clock In pulse animation, stat card counter animation, modal fade+scale
7. Add short descriptions on every page explaining what the page does
8. Add clear CTAs on every page (primary action buttons)
9. Create new reusable components: StatCard, PageHeader, EmptyState, SkeletonLoader, ConfirmModal, SearchBar, StatusBadge
10. Refactor: extract repeated patterns, use Svelte 5 runes consistently, Tailwind only (no inline styles)

## Design System
Colors: primary #1e3a5f, accent #3b82f6, success #10b981, warning #f59e0b, danger #ef4444
Font: Inter, bg: #f8fafc, surface: white, border: #e2e8f0
Cards: rounded-xl shadow-sm border border-slate-100 bg-white p-6

## Pages to Redesign
1. Login — split layout (branding left, form right), tagline, icons, loading state
2. Sidebar — logo, icons per menu, active state, user info bottom, logout button
3. Dashboard — greeting, 4 stat cards with icons+trends, bar chart, attendance table, CTAs
4. Attendance — real-time clock, big Clock In/Out buttons with animations, history table
5. Documents — drag-drop upload zone, file grid cards, status badges, HR review
6. Employees — search+filter, employee grid cards with avatars, add modal
7. AI HRD — generate form, loading animation, markdown result card, history accordion

## After redesign
Run: cd frontend && npm run build && npm run check
Both must pass with 0 errors.
