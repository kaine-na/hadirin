<script lang="ts">
	import { onMount } from 'svelte';
	import { BarChart3, Users, Clock, TrendingUp, AlertTriangle } from 'lucide-svelte';
	import { analyticsApi } from '$lib/api/analytics';
	import type { AttendanceSummary, DepartmentStat, TrendPoint, TopLateEmployee, AnalyticsFilter } from '$lib/api/analytics';
	import { toast } from '$lib/stores/toast.svelte';
	import ReportFilterBar from '$lib/components/ReportFilterBar.svelte';
	import AttendanceLineChart from '$lib/components/AttendanceLineChart.svelte';
	import DepartmentBarChart from '$lib/components/DepartmentBarChart.svelte';
	import StatusPieChart from '$lib/components/StatusPieChart.svelte';
	import ExportPDFButton from '$lib/components/ExportPDFButton.svelte';
	import AIExecutiveSummary from '$lib/components/AIExecutiveSummary.svelte';

	// State
	let loading = $state(false);
	let summary = $state<AttendanceSummary | null>(null);
	let deptStats = $state<DepartmentStat[]>([]);
	let trend = $state<TrendPoint[]>([]);
	let topLate = $state<TopLateEmployee[]>([]);

	// Filter state
	const today = new Date();
	const thirtyDaysAgo = new Date(today);
	thirtyDaysAgo.setDate(today.getDate() - 30);

	let filter = $state<AnalyticsFilter>({
		start_date: thirtyDaysAgo.toISOString().split('T')[0],
		end_date: today.toISOString().split('T')[0],
		department_id: ''
	});

	async function loadData() {
		loading = true;
		try {
			const f: AnalyticsFilter = {
				start_date: filter.start_date,
				end_date: filter.end_date,
				...(filter.department_id ? { department_id: filter.department_id } : {})
			};

			const [s, d, t, tl] = await Promise.all([
				analyticsApi.getAttendanceSummary(f),
				analyticsApi.getDepartmentStats(f),
				analyticsApi.getTrend(f),
				analyticsApi.getTopLateEmployees(f)
			]);

			summary = s;
			deptStats = d ?? [];
			trend = t ?? [];
			topLate = tl ?? [];
		} catch (err) {
			const msg = err instanceof Error ? err.message : 'Gagal memuat data analytics';
			toast.error(msg);
		} finally {
			loading = false;
		}
	}

	function handleFilter(start: string, end: string, dept: string) {
		filter = { start_date: start, end_date: end, department_id: dept };
		loadData();
	}

	onMount(() => {
		loadData();
	});

	function formatPct(val: number) {
		return val.toFixed(1) + '%';
	}
</script>

<svelte:head>
	<title>Laporan — Hadir</title>
</svelte:head>

<div class="space-y-6 p-6">
	<!-- Page header -->
	<div class="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
		<div class="flex items-center gap-3">
			<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-600 text-white shadow-sm">
				<BarChart3 size={20} />
			</div>
			<div>
				<h1 class="text-xl font-bold text-slate-900">Laporan & Analytics</h1>
				<p class="text-sm text-slate-500">Analisis kehadiran dan performa karyawan</p>
			</div>
		</div>
		<ExportPDFButton {filter} />
	</div>

	<!-- Filter bar -->
	<ReportFilterBar
		startDate={filter.start_date ?? ''}
		endDate={filter.end_date ?? ''}
		departmentId={filter.department_id ?? ''}
		onFilter={handleFilter}
	/>

	<!-- Summary cards -->
	{#if loading}
		<div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
			{#each Array(4) as _}
				<div class="h-24 animate-pulse rounded-xl bg-slate-100"></div>
			{/each}
		</div>
	{:else if summary}
		<div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
			<div class="rounded-xl bg-white p-5 shadow-md ring-1 ring-slate-100">
				<div class="mb-2 flex items-center gap-2 text-blue-600">
					<Users size={16} />
					<span class="text-xs font-medium text-slate-500">Total Karyawan</span>
				</div>
				<p class="text-2xl font-bold text-slate-900">{summary.total_employees}</p>
				<p class="text-xs text-slate-400">orang terdaftar</p>
			</div>

			<div class="rounded-xl bg-white p-5 shadow-md ring-1 ring-slate-100">
				<div class="mb-2 flex items-center gap-2 text-emerald-600">
					<TrendingUp size={16} />
					<span class="text-xs font-medium text-slate-500">Tingkat Kehadiran</span>
				</div>
				<p class="text-2xl font-bold text-slate-900">{formatPct(summary.attendance_rate)}</p>
				<p class="text-xs text-slate-400">{summary.total_present + summary.total_late} dari {summary.total_working_days} hari</p>
			</div>

			<div class="rounded-xl bg-white p-5 shadow-md ring-1 ring-slate-100">
				<div class="mb-2 flex items-center gap-2 text-amber-500">
					<Clock size={16} />
					<span class="text-xs font-medium text-slate-500">Tingkat Terlambat</span>
				</div>
				<p class="text-2xl font-bold text-slate-900">{formatPct(summary.lateness_rate)}</p>
				<p class="text-xs text-slate-400">{summary.total_late} hari terlambat</p>
			</div>

			<div class="rounded-xl bg-white p-5 shadow-md ring-1 ring-slate-100">
				<div class="mb-2 flex items-center gap-2 text-red-500">
					<AlertTriangle size={16} />
					<span class="text-xs font-medium text-slate-500">Total Alpha</span>
				</div>
				<p class="text-2xl font-bold text-slate-900">{summary.total_absent}</p>
				<p class="text-xs text-slate-400">hari tidak hadir tanpa keterangan</p>
			</div>
		</div>
	{/if}

	<!-- AI Executive Summary -->
	<AIExecutiveSummary {filter} />

	<!-- Charts grid -->
	<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
		{#if loading}
			<div class="h-80 animate-pulse rounded-xl bg-slate-100 lg:col-span-2"></div>
			<div class="h-80 animate-pulse rounded-xl bg-slate-100"></div>
			<div class="h-80 animate-pulse rounded-xl bg-slate-100"></div>
		{:else}
			<div class="lg:col-span-2">
				<AttendanceLineChart data={trend} />
			</div>
			<DepartmentBarChart data={deptStats} />
			<StatusPieChart data={summary} />
		{/if}
	</div>

	<!-- Top late employees table -->
	{#if !loading && topLate.length > 0}
		<div class="rounded-xl bg-white p-6 shadow-md">
			<h3 class="mb-1 text-sm font-semibold text-slate-900">Top 10 Karyawan Paling Sering Terlambat</h3>
			<p class="mb-4 text-xs text-slate-500">Karyawan dengan frekuensi keterlambatan tertinggi pada periode ini</p>
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-slate-100">
							<th class="pb-3 text-left text-xs font-semibold text-slate-500">#</th>
							<th class="pb-3 text-left text-xs font-semibold text-slate-500">Nama</th>
							<th class="pb-3 text-left text-xs font-semibold text-slate-500">Departemen</th>
							<th class="pb-3 text-right text-xs font-semibold text-slate-500">Terlambat</th>
							<th class="pb-3 text-right text-xs font-semibold text-slate-500">Alpha</th>
						</tr>
					</thead>
					<tbody>
						{#each topLate as emp, i}
							<tr class="border-b border-slate-50 hover:bg-slate-50">
								<td class="py-3 text-slate-400">{i + 1}</td>
								<td class="py-3 font-medium text-slate-900">{emp.name}</td>
								<td class="py-3 text-slate-600">{emp.department}</td>
								<td class="py-3 text-right">
									<span class="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-700">
										{emp.late_count}x
									</span>
								</td>
								<td class="py-3 text-right">
									{#if emp.absent_count > 0}
										<span class="rounded-full bg-red-100 px-2 py-0.5 text-xs font-medium text-red-700">
											{emp.absent_count}x
										</span>
									{:else}
										<span class="text-slate-400">—</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
