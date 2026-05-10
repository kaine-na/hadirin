<script lang="ts">
	import type { ChecklistItem, ChecklistStats } from '$lib/api/compliance';
	import { complianceApi } from '$lib/api/compliance';
	import { toast } from '$lib/stores/toast.svelte';
	import { formatDate } from '$lib/utils/format';
	import { CheckCircle2, Clock, AlertCircle, Calendar } from 'lucide-svelte';

	interface Props {
		items: ChecklistItem[];
		stats: ChecklistStats;
		onItemDone?: (item: ChecklistItem) => void;
	}

	let { items = $bindable([]), stats, onItemDone }: Props = $props();

	let loadingId = $state<string | null>(null);

	async function markDone(item: ChecklistItem) {
		if (item.status === 'done' || loadingId) return;
		loadingId = item.id;
		try {
			const updated = await complianceApi.markChecklistDone(item.id);
			// Update item di list
			const idx = items.findIndex((i) => i.id === item.id);
			if (idx !== -1) {
				items[idx] = updated;
			}
			toast.success(`"${item.title}" ditandai selesai`);
			onItemDone?.(updated);
		} catch {
			toast.error('Gagal menandai selesai');
		} finally {
			loadingId = null;
		}
	}

	function getStatusIcon(status: string) {
		switch (status) {
			case 'done':
				return CheckCircle2;
			case 'overdue':
				return AlertCircle;
			default:
				return Clock;
		}
	}

	function getStatusColor(status: string, daysUntil: number) {
		if (status === 'done') return 'text-emerald-600';
		if (status === 'overdue') return 'text-red-600';
		if (daysUntil <= 3) return 'text-amber-600';
		return 'text-slate-500';
	}

	function getRowBg(status: string, daysUntil: number) {
		if (status === 'done') return 'bg-emerald-50/50';
		if (status === 'overdue') return 'bg-red-50/50';
		if (daysUntil <= 3) return 'bg-amber-50/50';
		return 'bg-white';
	}

	function getDeadlineBadge(status: string, daysUntil: number) {
		if (status === 'done') return { text: 'Selesai', cls: 'bg-emerald-100 text-emerald-700' };
		if (status === 'overdue') return { text: 'Terlambat', cls: 'bg-red-100 text-red-700' };
		if (daysUntil === 0) return { text: 'Hari ini', cls: 'bg-red-100 text-red-700' };
		if (daysUntil <= 3) return { text: `${daysUntil} hari lagi`, cls: 'bg-amber-100 text-amber-700' };
		return { text: `${daysUntil} hari lagi`, cls: 'bg-slate-100 text-slate-600' };
	}
</script>

<div class="rounded-xl border border-slate-200 bg-white shadow-sm">
	<!-- Header dengan statistik -->
	<div class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
		<div>
			<h3 class="text-sm font-semibold text-slate-900">Checklist Kepatuhan</h3>
			<p class="mt-0.5 text-xs text-slate-500">Kewajiban compliance bulanan</p>
		</div>
		<div class="flex items-center gap-3">
			<div class="flex items-center gap-1.5 text-xs">
				<span class="h-2 w-2 rounded-full bg-emerald-500"></span>
				<span class="text-slate-600">{stats.done} selesai</span>
			</div>
			<div class="flex items-center gap-1.5 text-xs">
				<span class="h-2 w-2 rounded-full bg-amber-500"></span>
				<span class="text-slate-600">{stats.pending} pending</span>
			</div>
			{#if stats.overdue > 0}
				<div class="flex items-center gap-1.5 text-xs">
					<span class="h-2 w-2 rounded-full bg-red-500"></span>
					<span class="text-slate-600">{stats.overdue} terlambat</span>
				</div>
			{/if}
		</div>
	</div>

	<!-- List item -->
	<div class="divide-y divide-slate-100">
		{#each items as item (item.id)}
			{@const StatusIcon = getStatusIcon(item.status)}
			{@const statusColor = getStatusColor(item.status, item.days_until)}
			{@const rowBg = getRowBg(item.status, item.days_until)}
			{@const badge = getDeadlineBadge(item.status, item.days_until)}

			<div class="flex items-start gap-4 px-5 py-4 transition-colors {rowBg}">
				<!-- Status icon -->
				<div class="mt-0.5 shrink-0 {statusColor}">
					<StatusIcon size={18} strokeWidth={1.75} />
				</div>

				<!-- Konten -->
				<div class="min-w-0 flex-1">
					<div class="flex items-start justify-between gap-3">
						<div>
							<p class="text-sm font-medium text-slate-900 {item.status === 'done' ? 'line-through opacity-60' : ''}">
								{item.title}
							</p>
							<p class="mt-0.5 text-xs text-slate-500">{item.description}</p>
						</div>
						<div class="flex shrink-0 items-center gap-2">
							<!-- Deadline badge -->
							<span class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium {badge.cls}">
								<Calendar size={10} />
								{badge.text}
							</span>
						</div>
					</div>

					<div class="mt-2 flex items-center justify-between">
						<p class="text-xs text-slate-400">
							Deadline: {formatDate(item.deadline)}
						</p>
						{#if item.status !== 'done'}
							<button
								type="button"
								onclick={() => markDone(item)}
								disabled={loadingId === item.id}
								class="inline-flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white transition-all hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
							>
								{#if loadingId === item.id}
									<span class="h-3 w-3 animate-spin rounded-full border border-white border-t-transparent"></span>
									Menyimpan...
								{:else}
									<CheckCircle2 size={12} />
									Tandai Selesai
								{/if}
							</button>
						{:else}
							<p class="text-xs text-emerald-600">
								Selesai {item.done_at ? formatDate(item.done_at) : ''}
							</p>
						{/if}
					</div>
				</div>
			</div>
		{/each}

		{#if items.length === 0}
			<div class="px-5 py-10 text-center">
				<p class="text-sm text-slate-500">Belum ada checklist untuk periode ini</p>
			</div>
		{/if}
	</div>
</div>
