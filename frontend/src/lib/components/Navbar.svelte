<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Menu, X, ChevronDown, LogOut, LayoutDashboard, Clock, Shield, FileText, Users, Sparkles } from 'lucide-svelte';
	import { auth, clearAuth, isHRorManager } from '$lib/stores/auth.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { authApi } from '$lib/api';
	import { roleLabel } from '$lib/utils/format';
	import NotificationBell from './NotificationBell.svelte';

	let menuOpen = $state(false);
	let mobileOpen = $state(false);

	const userInitial = $derived(auth.user?.name?.charAt(0)?.toUpperCase() ?? '?');
	const userName = $derived(auth.user?.name ?? '-');
	const userRole = $derived(roleLabel(auth.user?.role ?? ''));

	// Page title derived from current path
	const pageTitle = $derived(() => {
		const path = $page.url.pathname;
		if (path === '/dashboard' || path === '/') return 'Dashboard';
		if (path.startsWith('/attendance/manage')) return 'Kelola Absensi';
		if (path.startsWith('/attendance')) return 'Absensi';
		if (path.startsWith('/documents')) return 'Dokumen';
		if (path.startsWith('/employees')) return 'Karyawan';
		if (path.startsWith('/hr-ai')) return 'HR AI';
		return 'SaaS Karyawan';
	});

	interface MobileNavItem {
		href: string;
		label: string;
		icon: typeof LayoutDashboard;
		roles?: () => boolean;
	}

	const mobileItems: MobileNavItem[] = [
		{ href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/attendance', label: 'Absensi', icon: Clock },
		{
			href: '/attendance/manage',
			label: 'Kelola Absensi',
			icon: Shield,
			roles: () => isHRorManager()
		},
		{ href: '/documents', label: 'Dokumen', icon: FileText },
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
		}
	];

	const visibleMobileItems = $derived(mobileItems.filter((i) => !i.roles || i.roles()));

	function isMobileActive(href: string): boolean {
		const path = $page.url.pathname;
		if (href === '/dashboard') return path === '/dashboard' || path === '/';
		return path === href || path.startsWith(href + '/');
	}

	async function logout() {
		menuOpen = false;
		mobileOpen = false;
		try {
			await authApi.logout();
		} catch {
			// abaikan error, lanjut clear
		}
		clearAuth();
		toast.success('Berhasil logout');
		await goto('/login');
	}

	function onDocClick(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (!target.closest('[data-profile-menu]')) {
			menuOpen = false;
		}
	}

	function closeMobile() {
		mobileOpen = false;
	}
</script>

<svelte:window onclick={onDocClick} />

<header class="flex h-16 shrink-0 items-center justify-between border-b border-slate-200 bg-white px-4 shadow-sm lg:px-6">
	<!-- Left: hamburger (mobile) + page title -->
	<div class="flex items-center gap-3">
		<button
			type="button"
			class="flex h-9 w-9 items-center justify-center rounded-lg text-slate-600 transition-colors hover:bg-slate-100 lg:hidden"
			aria-label="Buka menu"
			onclick={() => (mobileOpen = !mobileOpen)}
		>
			{#if mobileOpen}
				<X size={20} strokeWidth={2} />
			{:else}
				<Menu size={20} strokeWidth={2} />
			{/if}
		</button>

		<div class="hidden lg:block">
			<p class="text-sm font-semibold text-slate-800">{pageTitle()}</p>
		</div>
		<div class="lg:hidden">
			<p class="text-sm font-semibold text-slate-800">{pageTitle()}</p>
		</div>
	</div>

	<!-- Right: notification bell + user dropdown -->
	<div class="flex items-center gap-2">
		<!-- Notification bell (komponen dengan dropdown) -->
		<NotificationBell />

		<!-- User dropdown -->
		<div class="relative" data-profile-menu>
			<button
				type="button"
				class="flex items-center gap-2.5 rounded-lg px-2 py-1.5 transition-colors hover:bg-slate-100"
				onclick={() => (menuOpen = !menuOpen)}
			>
				<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-semibold text-blue-700">
					{userInitial}
				</div>
				<div class="hidden text-left sm:block">
					<p class="text-sm font-medium leading-tight text-slate-900">{userName}</p>
					<p class="text-xs leading-tight text-slate-500">{userRole}</p>
				</div>
				<ChevronDown
					size={14}
					strokeWidth={2}
					class="hidden shrink-0 text-slate-400 transition-transform duration-150 sm:block {menuOpen ? 'rotate-180' : ''}"
				/>
			</button>

			{#if menuOpen}
				<div
					class="absolute right-0 top-12 z-30 w-52 overflow-hidden rounded-xl border border-slate-200 bg-white shadow-lg"
				>
					<!-- User info header -->
					<div class="border-b border-slate-100 px-4 py-3">
						<div class="flex items-center gap-3">
							<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-blue-100 text-sm font-semibold text-blue-700">
								{userInitial}
							</div>
							<div class="min-w-0">
								<p class="truncate text-sm font-medium text-slate-900">{userName}</p>
								<p class="truncate text-xs text-slate-500">{userRole}</p>
							</div>
						</div>
					</div>
					<!-- Logout -->
					<div class="p-1">
						<button
							type="button"
							class="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-red-50 hover:text-red-600"
							onclick={logout}
						>
							<LogOut size={15} strokeWidth={1.75} class="shrink-0" />
							Keluar
						</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
</header>

<!-- Mobile drawer overlay -->
{#if mobileOpen}
	<div class="fixed inset-0 z-40 lg:hidden">
		<!-- Backdrop -->
		<button
			type="button"
			class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm"
			aria-label="Tutup menu"
			onclick={closeMobile}
		></button>

		<!-- Drawer panel -->
		<div class="relative flex h-full w-64 flex-col bg-white shadow-xl">
			<!-- Drawer header -->
			<div class="flex h-16 shrink-0 items-center justify-between border-b border-slate-200 px-5">
				<div class="flex items-center gap-2.5">
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-600 text-white">
						<span class="text-xs font-bold">SK</span>
					</div>
					<div>
						<p class="text-sm font-semibold text-slate-900">SaaS Karyawan</p>
						<p class="text-xs text-slate-500">Mirava Office</p>
					</div>
				</div>
				<button
					type="button"
					class="flex h-8 w-8 items-center justify-center rounded-lg text-slate-500 hover:bg-slate-100"
					aria-label="Tutup"
					onclick={closeMobile}
				>
					<X size={18} strokeWidth={2} />
				</button>
			</div>

			<!-- Drawer nav -->
			<nav class="flex flex-1 flex-col gap-0.5 overflow-y-auto p-3">
				{#each visibleMobileItems as item}
					{@const active = isMobileActive(item.href)}
					{@const Icon = item.icon}
					<a
						href={item.href}
						onclick={closeMobile}
						class="group flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-150
							{active
							? 'bg-blue-50 text-blue-700'
							: 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'}"
					>
						<span class="flex h-5 w-5 shrink-0 items-center justify-center {active ? 'text-blue-600' : 'text-slate-400 group-hover:text-slate-600'}">
							<Icon size={18} strokeWidth={1.75} />
						</span>
						{item.label}
					</a>
				{/each}
			</nav>

			<!-- Drawer user + logout -->
			<div class="shrink-0 border-t border-slate-200 p-3">
				<div class="mb-1 flex items-center gap-3 px-3 py-2">
					<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-semibold text-blue-700">
						{userInitial}
					</div>
					<div class="min-w-0">
						<p class="truncate text-sm font-medium text-slate-900">{userName}</p>
						<p class="truncate text-xs text-slate-500">{userRole}</p>
					</div>
				</div>
				<button
					type="button"
					onclick={logout}
					class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-red-50 hover:text-red-600"
				>
					<LogOut size={16} strokeWidth={1.75} class="shrink-0" />
					Keluar
				</button>
			</div>
		</div>
	</div>
{/if}
