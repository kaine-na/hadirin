<script lang="ts">
	import type { LeaveBalance } from '$lib/api/leaves';
	import { CalendarOff, TrendingDown } from 'lucide-svelte';

	interface Props {
		balances: LeaveBalance[];
		loading?: boolean;
	}

	let { balances, loading = false }: Props = $props();

	// Tampilkan hanya cuti tahunan di widget dashboard
	const annualLeave = $derived(
		balances.find((b) => b.leave_type_name?.toLowerCase().includes('tahunan'))
	);
</script>

<div
	class="rounded-xl border border-slate-100 bg-white p-5 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:shadow-md"
>
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-teal-100 text-teal-600"
			>
				<CalendarOff size={18} strokeWidth={1.75} />
			</div>
			<div>
				<p class="text-sm font-medium text-slate-500">Sisa Cuti Tahunan</p>
				{#if loading}
					<div class="mt-1 h-6 w-16 animate-pulse rounded bg-slate-200"></div>
				{:else if annualLeave}
					<p class="text-2xl font-bold text-slate-900">
						{annualLeave.remaining_days}
						<span class="text-sm font-normal text-slate-500">/ {annualLeave.total_days} hari</span>
					</p>
				{:else}
					<p class="text-lg font-semibold text-slate-400">Belum diatur</p>
				{/if}
			</div>
		</div>
		{#if annualLeave && annualLeave.used_days > 0}
			<div class="flex items-center gap-1 text-xs text-slate-500">
				<TrendingDown size={14} strokeWidth={1.75} class="text-amber-500" />
				<span>{annualLeave.used_days} terpakai</span>
			</div>
		{/if}
	</div>

	{#if !loading && balances.length > 0}
		<div class="mt-4 space-y-2">
			{#each balances.slice(0, 3) as balance}
				<div class="flex items-center justify-between text-xs">
					<span class="text-slate-600">{balance.leave_type_name}</span>
					<div class="flex items-center gap-2">
						<div class="h-1.5 w-20 overflow-hidden rounded-full bg-slate-100">
							<div
								class="h-full rounded-full bg-teal-500 transition-all duration-500"
								style="width: {balance.total_days > 0
									? Math.min(100, (balance.used_days / balance.total_days) * 100)
									: 0}%"
							></div>
						</div>
						<span class="w-12 text-right font-medium text-slate-700">
							{balance.remaining_days}/{balance.total_days}
						</span>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<a
		href="/leaves"
		class="mt-4 flex items-center gap-1 text-xs font-medium text-teal-600 hover:text-teal-700"
	>
		Lihat semua cuti →
	</a>
</div>
