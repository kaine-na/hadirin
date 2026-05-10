<script lang="ts">
	import { onMount } from 'svelte';
	import { complianceApi } from '$lib/api/compliance';
	import type { ComplianceSummary, EmployeeTHR, HolidayInfo, ChecklistItem, ChecklistStats } from '$lib/api/compliance';
	import ComplianceStatusCard from '$lib/components/ComplianceStatusCard.svelte';
	import ComplianceChecklist from '$lib/components/ComplianceChecklist.svelte';
	import BPJSCalculator from '$lib/components/BPJSCalculator.svelte';
	import PPh21Summary from '$lib/components/PPh21Summary.svelte';
	import THRCalculator from '$lib/components/THRCalculator.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { formatCurrency } from '$lib/utils/format';
	import { ShieldCheck, RefreshCw, Calendar } from 'lucide-svelte';

	// State
	let loading = $state(true);
	let summary = $state<ComplianceSummary | null>(null);
	let thrEmployees = $state<EmployeeTHR[]>([]);
	let thrHolidays = $state<HolidayInfo[]>([]);
	let thrTotal = $state(0);
	let thrYear = $state(new Date().getFullYear());

	// Periode yang dipilih
	const now = new Date();
	let selectedMonth = $state(`${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`);

	// Tab aktif
	let activeTab = $state<'overview' | 'bpjs' | 'pph21' | 'thr'>('overview');

	// Checklist state (diambil dari summary)
	let checklistItems = $state<ChecklistItem[]>([]);
	let checklistStats = $state<ChecklistStats>({ pending: 0, done: 0, overdue: 0, total: 0 });

	// Sync checklist dari summary
	$effect(() => {
		if (summary) {
			checklistItems = summary.checklist.items ?? [];
			checklistStats = {
				pending: summary.checklist.pending ?? 0,
				done: summary.checklist.done ?? 0,
				overdue: summary.checklist.overdue ?? 0,
				total: summary.checklist.items?.length ?? 0
			};
		}
	});

	async function loadData() {
		loading = true;
		try {
			const [summaryRes, thrRes] = await Promise.all([
				complianceApi.getSummary(selectedMonth),
				complianceApi.getTHRCalculation(thrYear)
			]);
			summary = summaryRes;
			thrEmployees = thrRes.employees ?? [];
			thrHolidays = thrRes.holidays ?? [];
			thrTotal = thrRes.total_thr ?? 0;
		} catch (err) {
			toast.error('Gagal memuat data compliance');
		} finally {
			loading = false;
		}
	}

	function handleChecklistItemDone(item: ChecklistItem) {
		// Refresh summary setelah item ditandai selesai
		loadData();
	}

	onMount(() => {
		loadData();
	});

	const tabs = [
		{ id: 'overview', label: 'Overview' },
		{ id: 'bpjs', label: 'BPJS' },
		{ id: 'pph21', label: 'PPh 21' },
		{ id: 'thr', label: 'THR' }
	] as const;
</script>

<svelte:head>
	<title>Kepatuhan — Hadir</title>
</svelte:head>

<div class="min-h-full bg-slate-50">
	<!-- Page header -->
	<div class="border-b border-slate-200 bg-white px-6 py-5">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-100">
					<ShieldCheck size={18} class="text-blue-600" strokeWidth={1.75} />
				</div>
				<div>
					<h1 class="text-lg font-semibold text-slate-900">Kepatuhan</h1>
					<p class="text-xs text-slate-500">Compliance Engine — BPJS, PPh 21, THR</p>
				</div>
			</div>

			<div class="flex items-center gap-3">
				<!-- Pilih periode -->
				<div class="flex items-center gap-2">
					<Calendar size={14} class="text-slate-400" />
					<input
						type="month"
						bind:value={selectedMonth}
						onchange={loadData}
						class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm text-slate-700 outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
					/>
				</div>

				<button
					type="button"
					onclick={loadData}
					disabled={loading}
					class="inline-flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-600 transition-all hover:bg-slate-50 disabled:opacity-50"
				>
					<RefreshCw size={14} class={loading ? 'animate-spin' : ''} />
					Refresh
				</button>
			</div>
		</div>

		<!-- Tabs -->
		<div class="mt-4 flex gap-1">
			{#each tabs as tab}
				<button
					type="button"
					onclick={() => (activeTab = tab.id)}
					class="rounded-lg px-4 py-2 text-sm font-medium transition-all {activeTab === tab.id
						? 'bg-blue-600 text-white'
						: 'text-slate-600 hover:bg-slate-100'}"
				>
					{tab.label}
				</button>
			{/each}
		</div>
	</div>

	<div class="p-6">
		{#if loading && !summary}
			<!-- Skeleton loading -->
			<div class="space-y-4">
				{#each [1, 2, 3] as _}
					<div class="h-24 animate-pulse rounded-xl bg-slate-200"></div>
				{/each}
			</div>
		{:else if activeTab === 'overview'}
			<div class="space-y-5">
				<!-- Status card -->
				{#if summary}
					<ComplianceStatusCard
						status={summary.overall_status}
						title="Status Kepatuhan {selectedMonth}"
						description={summary.overall_status === 'green'
							? 'Semua kewajiban compliance bulan ini dalam kondisi baik'
							: summary.overall_status === 'yellow'
								? 'Ada kewajiban yang mendekati deadline, segera selesaikan'
								: 'Ada kewajiban yang sudah melewati deadline!'}
						pending={summary.checklist.pending}
						done={summary.checklist.done}
						overdue={summary.checklist.overdue}
					/>

					<!-- BPJS Summary card -->
					<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
						<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
							<p class="text-xs text-slate-500">Total Karyawan</p>
							<p class="mt-1 text-2xl font-bold text-slate-900">{summary.bpjs_summary.total_employees}</p>
						</div>
						<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
							<p class="text-xs text-slate-500">Total Gaji Bruto</p>
							<p class="mt-1 text-lg font-bold text-slate-900">{formatCurrency(summary.bpjs_summary.total_gross_salary)}</p>
						</div>
						<div class="rounded-xl border border-blue-100 bg-blue-50 p-4 shadow-sm">
							<p class="text-xs text-blue-600">Tanggungan BPJS Perusahaan</p>
							<p class="mt-1 text-lg font-bold text-blue-900">{formatCurrency(summary.bpjs_summary.total_company_contribution)}</p>
						</div>
						<div class="rounded-xl border border-amber-100 bg-amber-50 p-4 shadow-sm">
							<p class="text-xs text-amber-600">Potongan BPJS Karyawan</p>
							<p class="mt-1 text-lg font-bold text-amber-900">{formatCurrency(summary.bpjs_summary.total_employee_deduction)}</p>
						</div>
					</div>
				{/if}

				<!-- Checklist -->
				<ComplianceChecklist
					bind:items={checklistItems}
					stats={checklistStats}
					onItemDone={handleChecklistItemDone}
				/>
			</div>
		{:else if activeTab === 'bpjs'}
			<div class="max-w-2xl">
				<BPJSCalculator />
			</div>
		{:else if activeTab === 'pph21'}
			<div class="max-w-2xl">
				<PPh21Summary />
			</div>
		{:else if activeTab === 'thr'}
			<THRCalculator
				employees={thrEmployees}
				totalTHR={thrTotal}
				holidays={thrHolidays}
				year={thrYear}
			/>
		{/if}
	</div>
</div>
