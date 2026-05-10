<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		LayoutDashboard,
		Clock,
		Shield,
		FileText,
		Users,
		Sparkles,
		Building2,
		LogOut,
		ChevronRight,
		CalendarOff,
		BarChart3,
		ShieldCheck,
		AlertTriangle
	} from 'lucide-svelte';
	import { auth, clearAuth, isHRorManager } from '$lib/stores/auth.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { authApi } from '$lib/api';
	import { roleLabel } from '$lib/utils/format';

	interface NavItem {
		href: string;
		label: string;
		icon: typeof LayoutDashboard;
		roles?: () => boolean;
	}

	const items: NavItem[] = [
		{ href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/attendance', label: 'Absensi', icon: Clock },
		{
			href: '/attendance/manage',
			label: 'Kelola Absensi',
			icon: Shield,
			roles: () => isHRorManager()
		},
		{ href: '/documents', label: 'Dokumen', icon: FileText },
		{ href: '/leaves', label: 'Cuti', icon: CalendarOff },
		{
			href: '/leaves/manage',
			label: 'Kelola Cuti',
			icon: Shield,
			roles: () => isHRorManager()
		},
		{
			href: '/employees',
			label: 'Karyawan',
			icon: Users,
			roles: () => isHRorManager()
		},
		{
			href: '/hr-ai',
			label: 'HR AI',
			icon: Sparkles,
			roles: () => isHRorManager()
		},
		{
			href: '/reports',
			label: 'Laporan',
			icon: BarChart3,
			roles: () => isHRorManager()
		},
		{
			href: '/compliance',
			label: 'Kepatuhan',
			icon: ShieldCheck,
			roles: () => isHRorManager()
		},
		{
			href: '/fraud',
			label: 'Deteksi Fraud',
			icon: AlertTriangle,
			roles: () => isHRorManager()
		}
	];

	function isActive(href: string): boolean {
		const path = $page.url.pathname;
		if (href === '/dashboard') return path === '/dashboard' || path === '/';
		return path === href || path.startsWith(href + '/');
	}

	const visibleItems = $derived(items.filter((i) => !i.roles || i.roles()));

	const userInitial = $derived(auth.user?.name?.charAt(0)?.toUpperCase() ?? '?');
	const userName = $derived(auth.user?.name ?? '-');
	const userRole = $derived(roleLabel(auth.user?.role ?? ''));

	async function logout() {
		try {
			await authApi.logout();
		} catch {
			// abaikan error, lanjut clear
		}
		clearAuth();
		toast.success('Berhasil logout');
		await goto('/login');
	}
</script>

<aside class="hidden w-64 shrink-0 flex-col border-r border-slate-200 bg-white shadow-sm lg:flex" style="height: 100vh; position: sticky; top: 0;">
	<!-- Header / Logo -->
	<div class="flex h-16 shrink-0 items-center gap-3 border-b border-slate-200 px-5">
		<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-600 text-white shadow-sm">
			<Building2 size={18} strokeWidth={2} />
		</div>
		<div class="min-w-0">
			<p class="truncate text-sm font-semibold text-slate-900">Hadir</p>
			<p class="truncate text-xs text-slate-500">Mirava Office</p>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="flex flex-1 flex-col gap-0.5 overflow-y-auto p-3">
		{#each visibleItems as item}
			{#if true}
				{@const active = isActive(item.href)}
				{@const Icon = item.icon}
				<a
					href={item.href}
					class="group flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-150
						{active
						? 'border-r-2 border-blue-600 bg-blue-50 text-blue-700'
						: 'border-r-2 border-transparent text-slate-600 hover:translate-x-0.5 hover:bg-slate-50 hover:text-slate-900'}"
				>
					<span class="flex h-5 w-5 shrink-0 items-center justify-center {active ? 'text-blue-600' : 'text-slate-400 group-hover:text-slate-600'}">
						<Icon size={18} strokeWidth={1.75} />
					</span>
					<span class="flex-1 truncate">{item.label}</span>
					{#if active}
						<ChevronRight size={14} strokeWidth={2} class="shrink-0 text-blue-500" />
					{/if}
				</a>
			{/if}
		{/each}
	</nav>

	<!-- User info + Logout -->
	<div class="shrink-0 border-t border-slate-200 p-3">
		<div class="mb-2 flex items-center gap-3 rounded-lg px-3 py-2.5">
			<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-semibold text-blue-700">
				{userInitial}
			</div>
			<div class="min-w-0 flex-1">
				<p class="truncate text-sm font-medium text-slate-900">{userName}</p>
				<p class="truncate text-xs text-slate-500">{userRole}</p>
			</div>
		</div>
		<button
			type="button"
			onclick={logout}
			class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-slate-600 transition-all duration-150 hover:bg-red-50 hover:text-red-600"
		>
			<LogOut size={16} strokeWidth={1.75} class="shrink-0" />
			<span>Keluar</span>
		</button>
	</div>
</aside>
