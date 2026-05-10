<script lang="ts">
	import { onMount } from 'svelte';
	import { fraudApi } from '$lib/api/fraud';
	import type { FraudLog, FraudSummary } from '$lib/api/fraud';
	import { toast } from '$lib/stores/toast.svelte';
	import { isHR } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import FraudLogCard from '$lib/components/FraudLogCard.svelte';
	import FraudSummaryWidget from '$lib/components/FraudSummaryWidget.svelte';
	import SkeletonLoader from '$lib/components/ui/SkeletonLoader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { AlertTriangle, RefreshCw, Filter } from 'lucide-svelte';

	let logs = $state<FraudLog[]>([]);
	let summary = $state<FraudSummary | null>(null);
	let loadingLogs = $state(true);
	let loadingSummary = $state(true);
	let statusFilter = $state('');
	let page = $state(1);
	let total = $state(0);
	const pageSize = 10;

	const totalPages = $derived(Math.ceil(total / pageSize));

	async function loadLogs() {
		loadingLogs = true;
		try {
			const res = await fraudApi.listLogs({
				status: statusFilter || undefined,
				page,
				page_size: pageSize
			});
			logs = res.items ?? [];
			total = res.total ?? 0;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal memuat fraud logs');
		} finally {
			loadingLogs = false;
		}
	}

	async function loadSummary() {
		loadingSummary = true;
		try {
			summary = await fraudApi.getSummary();
		} catch {
			summary = null;
		} finally {
			loadingSummary = false;
		}
	}

	function applyFilter() {
		page = 1;
		loadLogs();
	}

	function onLogUpdate() {
		loadLogs();
		loadSummary();
	}

	onMount(async () => {
		// Redirect jika bukan HR
		if (!isHR()) {
			await goto('/dashboard');
			return;
		}
		await Promise.all([loadLogs(), loadSummary()]);
	});
</script>

<svelte:head>
	<title>Deteksi Fraud — Hadir</title>
</svelte:head>

<div class="space-y-6">
	<PageHeader
		title="Deteksi Fraud Absensi"
		description="Monitor dan tinjau aktivitas absensi yang mencurigakan."
	/>

	<!-- Layout: Summary widget + Logs -->
	<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
		<!-- Summary Widget (sidebar kiri) -->
		<div class="lg:col-span-1">
			<FraudSummaryWidget {summary} loading={loadingSummary} />
		</div>

		<!-- Fraud Logs (konten utama) -->
		<div class="lg:col-span-2 space-y-4">
			<!-- Filter bar -->
			<div class="flex flex-wrap items-center gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
				<div class="flex items-center gap-2 text-slate-500">
					<Filter size={16} strokeWidth={2} />
					<span class="text-sm font-medium">Filter:</span>
				</div>
				<div class="flex flex-wrap gap-2">
					{#each [
						{ value: '', label: 'Semua' },
						{ value: 'pending', label: 'Pending' },
						{ value: 'confirmed', label: 'Dikonfirmasi' },
						{ value: 'dismissed', label: 'Diabaikan' }
					] as filter}
						<button
							type="button"
							onclick={() => { statusFilter = filter.value; applyFilter(); }}
							class="rounded-full px-3 py-1.5 text-xs font-medium transition-colors
								{statusFilter === filter.value
									? 'bg-blue-600 text-white'
									: 'border border-slate-200 bg-white text-slate-600 hover:bg-slate-50'}"
						>
							{filter.label}
						</button>
					{/each}
				</div>
				<button
					type="button"
					onclick={loadLogs}
					disabled={loadingLogs}
					class="ml-auto flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-50"
				>
					<RefreshCw size={13} strokeWidth={2} class={loadingLogs ? 'animate-spin' : ''} />
					Refresh
				</button>
			</div>

			<!-- Daftar logs -->
			{#if loadingLogs}
				<div class="space-y-3">
					{#each Array(3) as _}
						<SkeletonLoader rows={3} type="card" />
					{/each}
				</div>
			{:else if logs.length === 0}
				<EmptyState
					icon={AlertTriangle}
					title="Tidak ada fraud log"
					description={statusFilter
						? `Tidak ada fraud log dengan status "${statusFilter}"`
						: 'Belum ada aktivitas mencurigakan yang terdeteksi'}
				/>
			{:else}
				<div class="space-y-3">
					{#each logs as log (log.id)}
						<FraudLogCard {log} onUpdate={onLogUpdate} />
					{/each}
				</div>

				<!-- Pagination -->
				{#if totalPages > 1}
					<div class="flex items-center justify-between rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
						<p class="text-sm text-slate-500">
							Menampilkan {(page - 1) * pageSize + 1}–{Math.min(page * pageSize, total)} dari {total} log
						</p>
						<div class="flex gap-2">
							<button
								type="button"
								onclick={() => { page--; loadLogs(); }}
								disabled={page <= 1}
								class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-40"
							>
								Sebelumnya
							</button>
							<button
								type="button"
								onclick={() => { page++; loadLogs(); }}
								disabled={page >= totalPages}
								class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-40"
							>
								Berikutnya
							</button>
						</div>
					</div>
				{/if}
			{/if}
		</div>
	</div>
</div>
