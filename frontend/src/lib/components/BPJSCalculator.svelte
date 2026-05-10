<script lang="ts">
	import type { BPJSResult } from '$lib/api/compliance';
	import { complianceApi } from '$lib/api/compliance';
	import { toast } from '$lib/stores/toast.svelte';
	import { formatCurrency } from '$lib/utils/format';
	import { Calculator, ChevronDown, ChevronUp } from 'lucide-svelte';

	let grossSalary = $state(5000000);
	let result = $state<BPJSResult | null>(null);
	let loading = $state(false);
	let showDetail = $state(false);

	async function calculate() {
		if (grossSalary <= 0) {
			toast.error('Masukkan gaji bruto yang valid');
			return;
		}
		loading = true;
		try {
			const res = await complianceApi.getBPJSCalculation({ gross_salary: grossSalary });
			result = res.result;
		} catch {
			toast.error('Gagal menghitung BPJS');
		} finally {
			loading = false;
		}
	}

	function formatPct(rate: number) {
		return (rate * 100).toFixed(2) + '%';
	}
</script>

<div class="rounded-xl border border-slate-200 bg-white shadow-sm">
	<div class="border-b border-slate-100 px-5 py-4">
		<div class="flex items-center gap-2">
			<Calculator size={16} class="text-blue-600" strokeWidth={1.75} />
			<h3 class="text-sm font-semibold text-slate-900">Kalkulator BPJS</h3>
		</div>
		<p class="mt-0.5 text-xs text-slate-500">Hitung iuran BPJS berdasarkan gaji bruto</p>
	</div>

	<div class="p-5">
		<!-- Input gaji -->
		<div class="mb-4">
			<label for="gross-salary" class="mb-1.5 block text-xs font-medium text-slate-700">
				Gaji Bruto Bulanan
			</label>
			<div class="flex gap-2">
				<div class="relative flex-1">
					<span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-slate-400">Rp</span>
					<input
						id="gross-salary"
						type="number"
						bind:value={grossSalary}
						min="0"
						step="100000"
						class="w-full rounded-lg border border-slate-200 py-2.5 pl-10 pr-3 text-sm text-slate-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
						placeholder="5000000"
					/>
				</div>
				<button
					type="button"
					onclick={calculate}
					disabled={loading}
					class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition-all hover:bg-blue-700 disabled:opacity-50"
				>
					{#if loading}
						<span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
					{:else}
						<Calculator size={14} />
					{/if}
					Hitung
				</button>
			</div>
		</div>

		{#if result}
			<!-- Ringkasan -->
			<div class="mb-4 grid grid-cols-3 gap-3">
				<div class="rounded-lg bg-blue-50 p-3 text-center">
					<p class="text-xs text-blue-600">Tanggungan Perusahaan</p>
					<p class="mt-1 text-sm font-semibold text-blue-900">{formatCurrency(result.total_company_contribution)}</p>
				</div>
				<div class="rounded-lg bg-amber-50 p-3 text-center">
					<p class="text-xs text-amber-600">Potongan Karyawan</p>
					<p class="mt-1 text-sm font-semibold text-amber-900">{formatCurrency(result.total_employee_contribution)}</p>
				</div>
				<div class="rounded-lg bg-emerald-50 p-3 text-center">
					<p class="text-xs text-emerald-600">Take Home Pay</p>
					<p class="mt-1 text-sm font-semibold text-emerald-900">{formatCurrency(result.take_home_pay)}</p>
				</div>
			</div>

			<!-- Toggle detail -->
			<button
				type="button"
				onclick={() => (showDetail = !showDetail)}
				class="flex w-full items-center justify-between rounded-lg border border-slate-200 px-4 py-2.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50"
			>
				<span>Lihat Rincian Komponen</span>
				{#if showDetail}
					<ChevronUp size={14} />
				{:else}
					<ChevronDown size={14} />
				{/if}
			</button>

			{#if showDetail}
				<div class="mt-3 space-y-2">
					<!-- BPJS Kesehatan -->
					<div class="rounded-lg border border-slate-100 p-3">
						<p class="mb-2 text-xs font-semibold text-slate-700">BPJS Kesehatan</p>
						<p class="mb-1 text-xs text-slate-400">Basis gaji: {formatCurrency(result.kes_base_salary)} (max Rp 12 juta)</p>
						<div class="space-y-1">
							<div class="flex justify-between text-xs">
								<span class="text-slate-500">Perusahaan ({formatPct(result.kes_company_rate)})</span>
								<span class="font-medium text-slate-700">{formatCurrency(result.kes_company)}</span>
							</div>
							<div class="flex justify-between text-xs">
								<span class="text-slate-500">Karyawan ({formatPct(result.kes_employee_rate)})</span>
								<span class="font-medium text-slate-700">{formatCurrency(result.kes_employee)}</span>
							</div>
						</div>
					</div>

					<!-- BPJS TK JHT -->
					<div class="rounded-lg border border-slate-100 p-3">
						<p class="mb-2 text-xs font-semibold text-slate-700">BPJS TK - JHT</p>
						<div class="space-y-1">
							<div class="flex justify-between text-xs">
								<span class="text-slate-500">Perusahaan ({formatPct(result.jht_company_rate)})</span>
								<span class="font-medium text-slate-700">{formatCurrency(result.jht_company)}</span>
							</div>
							<div class="flex justify-between text-xs">
								<span class="text-slate-500">Karyawan ({formatPct(result.jht_employee_rate)})</span>
								<span class="font-medium text-slate-700">{formatCurrency(result.jht_employee)}</span>
							</div>
						</div>
					</div>

					<!-- BPJS TK JP -->
					<div class="rounded-lg border border-slate-100 p-3">
						<p class="mb-2 text-xs font-semibold text-slate-700">BPJS TK - JP</p>
						<p class="mb-1 text-xs text-slate-400">Basis gaji: {formatCurrency(result.jp_base_salary)} (max Rp 9.559.600)</p>
						<div class="space-y-1">
							<div class="flex justify-between text-xs">
								<span class="text-slate-500">Perusahaan ({formatPct(result.jp_company_rate)})</span>
								<span class="font-medium text-slate-700">{formatCurrency(result.jp_company)}</span>
							</div>
							<div class="flex justify-between text-xs">
								<span class="text-slate-500">Karyawan ({formatPct(result.jp_employee_rate)})</span>
								<span class="font-medium text-slate-700">{formatCurrency(result.jp_employee)}</span>
							</div>
						</div>
					</div>

					<!-- JKK + JKM -->
					<div class="rounded-lg border border-slate-100 p-3">
						<p class="mb-2 text-xs font-semibold text-slate-700">BPJS TK - JKK & JKM</p>
						<div class="space-y-1">
							<div class="flex justify-between text-xs">
								<span class="text-slate-500">JKK Perusahaan ({formatPct(result.jkk_company_rate)})</span>
								<span class="font-medium text-slate-700">{formatCurrency(result.jkk_company)}</span>
							</div>
							<div class="flex justify-between text-xs">
								<span class="text-slate-500">JKM Perusahaan ({formatPct(result.jkm_company_rate)})</span>
								<span class="font-medium text-slate-700">{formatCurrency(result.jkm_company)}</span>
							</div>
						</div>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>
