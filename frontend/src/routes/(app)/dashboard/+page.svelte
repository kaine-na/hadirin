<script lang="ts">
	import { onMount } from 'svelte';
	import { attendanceApi, employeesApi, documentsApi, leavesApi } from '$lib/api';
	import { auth, isHRorManager } from '$lib/stores/auth.svelte';
	import { formatTime, formatDate, roleLabel, statusLabel, statusColor } from '$lib/utils/format';
	import Badge from '$lib/components/Badge.svelte';
	import StatCard from '$lib/components/ui/StatCard.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import SkeletonLoader from '$lib/components/ui/SkeletonLoader.svelte';
	import LeaveBalanceWidget from '$lib/components/LeaveBalanceWidget.svelte';
	import type { LeaveBalance } from '$lib/api/leaves';
	import {
		Users,
		Clock,
		FileText,
		Sparkles,
		CheckCircle,
		AlertCircle,
		ArrowRight,
		LayoutDashboard,
		CalendarOff
	} from 'lucide-svelte';
	import type { Attendance } from '$lib/types';

	let today = $state<Attendance | null>(null);
	let loadingToday = $state(true);

	let stats = $state({
		employees: 0,
		documents: 0,
		myAttendanceMonth: 0
	});
	let loadingStats = $state(true);

	let leaveBalances = $state<LeaveBalance[]>([]);
	let loadingLeave = $state(true);

	onMount(async () => {
		// Status hari ini — untuk semua role
		try {
			today = await attendanceApi.today();
		} catch {
			today = null;
		} finally {
			loadingToday = false;
		}

		// Saldo cuti
		try {
			leaveBalances = await leavesApi.getMyBalance();
		} catch {
			// abaikan
		} finally {
			loadingLeave = false;
		}

		// Stats
		try {
			if (isHRorManager()) {
				const [emp, docs] = await Promise.all([
					employeesApi.list({ page: 1, page_size: 1 }),
					documentsApi.list({ page: 1, page_size: 1 })
				]);
				stats.employees = emp.total;
				stats.documents = docs.total;
			}
			const meAtt = await attendanceApi.me({ page: 1, page_size: 1 });
			stats.myAttendanceMonth = meAtt.total;
		} catch {
			// abaikan
		} finally {
			loadingStats = false;
		}
	});

	const nowStr = $derived(
		new Date().toLocaleDateString('id-ID', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		})
	);

	const greeting = $derived(() => {
		const hour = new Date().getHours();
		if (hour < 12) return 'Selamat pagi';
		if (hour < 15) return 'Selamat siang';
		if (hour < 18) return 'Selamat sore';
		return 'Selamat malam';
	});
</script>

<svelte:head>
	<title>Dashboard — SaaS Karyawan</title>
</svelte:head>

<div class="space-y-6">
	<!-- Welcome Header -->
	<div class="flex flex-wrap items-end justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-slate-900">
				{greeting()}, {auth.user?.name?.split(' ')[0] ?? 'Karyawan'} 👋
			</h1>
			<p class="mt-1 text-sm text-slate-500">{nowStr} · Ringkasan aktivitas karyawan hari ini</p>
		</div>
		<div class="flex items-center gap-2">
			<Badge color="primary">{roleLabel(auth.user?.role ?? '')}</Badge>
			{#if auth.user?.department}
				<Badge color="slate">{auth.user.department}</Badge>
			{/if}
		</div>
	</div>

	<!-- Status Absensi Hari Ini -->
	<div
		class="rounded-xl border border-slate-100 bg-white p-6 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:shadow-md"
	>
		<div class="flex flex-wrap items-center justify-between gap-4">
			<div class="flex items-center gap-4">
				<div
					class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl
					{today
						? today.clock_out
							? 'bg-slate-100 text-slate-500'
							: 'bg-green-100 text-green-600'
						: 'bg-amber-100 text-amber-600'}"
				>
					{#if today?.clock_out}
						<CheckCircle size={22} strokeWidth={1.75} />
					{:else if today}
						<Clock size={22} strokeWidth={1.75} />
					{:else}
						<AlertCircle size={22} strokeWidth={1.75} />
					{/if}
				</div>
				<div>
					<p class="text-sm font-medium text-slate-500">Status Absensi Hari Ini</p>
					{#if loadingToday}
						<p class="mt-0.5 text-lg font-semibold text-slate-400">Memuat...</p>
					{:else if today}
						<div class="mt-1 flex flex-wrap items-center gap-2">
							<Badge color={statusColor(today.status)}>{statusLabel(today.status)}</Badge>
							<span class="text-sm text-slate-600">
								Masuk: <span class="font-medium text-slate-900">{formatTime(today.clock_in)}</span>
								{#if today.clock_out}
									· Pulang: <span class="font-medium text-slate-900"
										>{formatTime(today.clock_out)}</span
									>
								{/if}
							</span>
						</div>
					{:else}
						<p class="mt-0.5 text-lg font-semibold text-slate-900">Belum absen hari ini</p>
					{/if}
				</div>
			</div>
			<a
				href="/attendance"
				class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-all duration-150 hover:bg-blue-700 hover:shadow-md active:scale-95"
			>
				{today?.clock_out ? 'Lihat Rekap' : 'Ke Halaman Absensi'}
				<ArrowRight size={14} strokeWidth={2} />
			</a>
		</div>
	</div>

	<!-- Stat Cards -->
	{#if loadingStats}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			<SkeletonLoader rows={3} />
			{#if isHRorManager()}
				<SkeletonLoader rows={3} />
				<SkeletonLoader rows={3} />
			{/if}
		</div>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			<StatCard
				title="Rekap Absensi Saya"
				value={stats.myAttendanceMonth}
				icon={Clock}
				color="blue"
				trend="total catatan"
			/>

			{#if isHRorManager()}
				<StatCard
					title="Total Karyawan"
					value={stats.employees}
					icon={Users}
					color="green"
					trend="terdaftar di sistem"
				/>

				<StatCard
					title="Total Dokumen"
					value={stats.documents}
					icon={FileText}
					color="purple"
					trend="diunggah karyawan"
				/>
			{/if}
		</div>
	{/if}

	<!-- Leave Balance Widget -->
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		<LeaveBalanceWidget balances={leaveBalances} loading={loadingLeave} />
	</div>

	<!-- Quick Links -->
	<div>
		<h2 class="text-xs font-semibold uppercase tracking-wider text-slate-400">Akses Cepat</h2>
		<div class="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
			<a
				href="/attendance"
				class="group flex items-center gap-3 rounded-xl border border-slate-100 bg-white p-4 shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-blue-200 hover:shadow-md"
			>
				<div
					class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-100 text-blue-600 transition-colors group-hover:bg-blue-600 group-hover:text-white"
				>
					<Clock size={18} strokeWidth={1.75} />
				</div>
				<div class="min-w-0">
					<p class="text-sm font-semibold text-slate-900">Absensi</p>
					<p class="text-xs text-slate-500">Clock in & rekap</p>
				</div>
				<ArrowRight
					size={14}
					strokeWidth={2}
					class="ml-auto shrink-0 text-slate-300 transition-colors group-hover:text-blue-500"
				/>
			</a>

			<a
				href="/documents/upload"
				class="group flex items-center gap-3 rounded-xl border border-slate-100 bg-white p-4 shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-blue-200 hover:shadow-md"
			>
				<div
					class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-indigo-100 text-indigo-600 transition-colors group-hover:bg-indigo-600 group-hover:text-white"
				>
					<FileText size={18} strokeWidth={1.75} />
				</div>
				<div class="min-w-0">
					<p class="text-sm font-semibold text-slate-900">Upload Dokumen</p>
					<p class="text-xs text-slate-500">Laporan kerja</p>
				</div>
				<ArrowRight
					size={14}
					strokeWidth={2}
					class="ml-auto shrink-0 text-slate-300 transition-colors group-hover:text-indigo-500"
				/>
			</a>

			{#if isHRorManager()}
				<a
					href="/employees"
					class="group flex items-center gap-3 rounded-xl border border-slate-100 bg-white p-4 shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-blue-200 hover:shadow-md"
				>
					<div
						class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-emerald-100 text-emerald-600 transition-colors group-hover:bg-emerald-600 group-hover:text-white"
					>
						<Users size={18} strokeWidth={1.75} />
					</div>
					<div class="min-w-0">
						<p class="text-sm font-semibold text-slate-900">Karyawan</p>
						<p class="text-xs text-slate-500">Kelola data</p>
					</div>
					<ArrowRight
						size={14}
						strokeWidth={2}
						class="ml-auto shrink-0 text-slate-300 transition-colors group-hover:text-emerald-500"
					/>
				</a>

				<a
					href="/hr-ai"
					class="group flex items-center gap-3 rounded-xl border border-slate-100 bg-white p-4 shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-blue-200 hover:shadow-md"
				>
					<div
						class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-violet-100 text-violet-600 transition-colors group-hover:bg-violet-600 group-hover:text-white"
					>
						<Sparkles size={18} strokeWidth={1.75} />
					</div>
					<div class="min-w-0">
						<p class="text-sm font-semibold text-slate-900">HR AI</p>
						<p class="text-xs text-slate-500">Analisis kinerja</p>
					</div>
					<ArrowRight
						size={14}
						strokeWidth={2}
						class="ml-auto shrink-0 text-slate-300 transition-colors group-hover:text-violet-500"
					/>
				</a>
			{/if}
		</div>
	</div>

	<!-- CTA Section (HR only) -->
	{#if isHRorManager()}
		<div
			class="rounded-xl border border-blue-100 bg-gradient-to-r from-blue-50 to-indigo-50 p-6 shadow-sm"
		>
			<div class="flex flex-wrap items-center justify-between gap-4">
				<div class="flex items-center gap-4">
					<div
						class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-blue-600 text-white shadow-sm"
					>
						<Sparkles size={22} strokeWidth={1.75} />
					</div>
					<div>
						<p class="font-semibold text-slate-900">Generate Laporan AI</p>
						<p class="text-sm text-slate-500">
							Analisis kinerja karyawan menggunakan kecerdasan buatan
						</p>
					</div>
				</div>
				<a
					href="/hr-ai"
					class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 text-sm font-medium text-white shadow-sm transition-all duration-150 hover:bg-blue-700 hover:shadow-md active:scale-95"
				>
					Mulai Analisis
					<ArrowRight size={14} strokeWidth={2} />
				</a>
			</div>
		</div>
	{/if}
</div>
