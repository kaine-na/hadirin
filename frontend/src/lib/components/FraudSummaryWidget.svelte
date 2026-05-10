<script lang="ts">
	import { AlertTriangle, CheckCircle, Clock, TrendingUp } from 'lucide-svelte';
	import type { FraudSummary } from '$lib/api/fraud';

	interface Props {
		summary: FraudSummary | null;
		loading?: boolean;
	}

	let { summary, loading = false }: Props = $props();

	const hasActiveFraud = $derived((summary?.pending_logs ?? 0) > 0 || (summary?.confirmed_logs ?? 0) > 0);

	const fraudTypeLabel: Record<string, string> = {
		gps_accuracy: 'GPS Buruk',
		mock_location: 'Lokasi Palsu',
		velocity_check: 'Velocity',
		anomaly_time: 'Waktu Aneh',
		anomaly_location: 'Lokasi Aneh',
		anomaly_device: 'Device Baru',
		liveness_fail: 'Selfie Gagal'
	};
</script>

<div
	class="overflow-hidden rounded-xl border shadow-sm transition-shadow hover:shadow-md
		{hasActiveFraud ? 'border-red-200 bg-red-50' : 'border-slate-200 bg-white'}"
>
	<!-- Header -->
	<div class="flex items-center justify-between px-5 py-4 border-b {hasActiveFraud ? 'border-red-200' : 'border-slate-100'}">
		<div class="flex items-center gap-2">
			<AlertTriangle
				size={18}
				strokeWidth={2}
				class={hasActiveFraud ? 'text-red-600' : 'text-slate-400'}
			/>
			<h3 class="text-sm font-semibold {hasActiveFraud ? 'text-red-900' : 'text-slate-700'}">
				Deteksi Fraud Bulan Ini
			</h3>
		</div>
		{#if hasActiveFraud}
			<span class="inline-flex items-center gap-1 rounded-full bg-red-600 px-2.5 py-0.5 text-xs font-semibold text-white">
				<span class="h-1.5 w-1.5 rounded-full bg-white animate-pulse"></span>
				Aktif
			</span>
		{:else}
			<span class="inline-flex items-center gap-1 rounded-full bg-emerald-100 px-2.5 py-0.5 text-xs font-medium text-emerald-700">
				<CheckCircle size={12} strokeWidth={2} />
				Aman
			</span>
		{/if}
	</div>

	<!-- Konten -->
	{#if loading}
		<div class="flex items-center justify-center py-8">
			<div class="h-6 w-6 animate-spin rounded-full border-2 border-slate-300 border-t-blue-600"></div>
		</div>
	{:else if summary}
		<div class="p-5 space-y-4">
			<!-- Angka besar total -->
			<div class="flex items-end gap-4">
				<div class="text-center">
					<p
						class="text-4xl font-bold tabular-nums {hasActiveFraud ? 'text-red-600' : 'text-slate-800'}"
					>
						{summary.total_logs}
					</p>
					<p class="text-xs text-slate-500 mt-0.5">Total Log</p>
				</div>
				<div class="flex-1 grid grid-cols-3 gap-2">
					<div class="rounded-lg bg-amber-50 border border-amber-100 px-3 py-2 text-center">
						<p class="text-lg font-bold text-amber-700">{summary.pending_logs}</p>
						<p class="text-xs text-amber-600">Pending</p>
					</div>
					<div class="rounded-lg bg-red-50 border border-red-100 px-3 py-2 text-center">
						<p class="text-lg font-bold text-red-700">{summary.confirmed_logs}</p>
						<p class="text-xs text-red-600">Dikonfirmasi</p>
					</div>
					<div class="rounded-lg bg-slate-50 border border-slate-100 px-3 py-2 text-center">
						<p class="text-lg font-bold text-slate-600">{summary.dismissed_logs}</p>
						<p class="text-xs text-slate-500">Diabaikan</p>
					</div>
				</div>
			</div>

			<!-- Tipe fraud terbanyak -->
			{#if Object.keys(summary.by_type).length > 0}
				<div>
					<p class="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">
						Tipe Fraud
					</p>
					<div class="space-y-1.5">
						{#each Object.entries(summary.by_type).sort((a, b) => b[1] - a[1]).slice(0, 3) as [type, count]}
							<div class="flex items-center justify-between">
								<span class="text-xs text-slate-600">{fraudTypeLabel[type] ?? type}</span>
								<div class="flex items-center gap-2">
									<div class="h-1.5 rounded-full bg-red-200" style="width: {Math.max(20, count * 20)}px">
										<div class="h-full rounded-full bg-red-500" style="width: 100%"></div>
									</div>
									<span class="text-xs font-semibold text-slate-700 w-4 text-right">{count}</span>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Top employees -->
			{#if summary.top_employees?.length > 0}
				<div>
					<p class="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">
						Karyawan Teratas
					</p>
					<div class="space-y-1.5">
						{#each summary.top_employees.slice(0, 3) as emp}
							<div class="flex items-center justify-between">
								<span class="text-xs text-slate-700 truncate max-w-[140px]">{emp.employee_name}</span>
								<span class="text-xs font-semibold text-red-600">{emp.fraud_count}x</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<div class="flex flex-col items-center justify-center py-8 text-center">
			<CheckCircle size={32} strokeWidth={1.5} class="text-emerald-400 mb-2" />
			<p class="text-sm text-slate-500">Tidak ada data fraud</p>
		</div>
	{/if}
</div>
