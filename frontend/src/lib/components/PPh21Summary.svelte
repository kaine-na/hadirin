<script lang="ts">
	import type { PPh21Result } from '$lib/api/compliance';
	import { complianceApi } from '$lib/api/compliance';
	import { toast } from '$lib/stores/toast.svelte';
	import { formatCurrency } from '$lib/utils/format';
	import { Receipt, Info } from 'lucide-svelte';

	let grossSalary = $state(5000000);
	let marital = $state<'TK' | 'K'>('TK');
	let dependents = $state(0);
	let result = $state<PPh21Result | null>(null);
	let loading = $state(false);

	async function calculate() {
		if (grossSalary <= 0) {
			toast.error('Masukkan gaji bruto yang valid');
			return;
		}
		loading = true;
		try {
			const res = await complianceApi.getPPh21Calculation({
				gross_salary: grossSalary,
				marital,
				dependents
			});
			result = res.result;
		} catch {
			toast.error('Gagal menghitung PPh 21');
		} finally {
			loading = false;
		}
	}

	const categoryLabel: Record<string, string> = {
		A: 'Kategori A (TK/0)',
		B: 'Kategori B (K/0, TK/1)',
		C: 'Kategori C (K/1, K/2, TK/2, TK/3)'
	};
</script>

<div class="rounded-xl border border-slate-200 bg-white shadow-sm">
	<div class="border-b border-slate-100 px-5 py-4">
		<div class="flex items-center gap-2">
			<Receipt size={16} class="text-purple-600" strokeWidth={1.75} />
			<h3 class="text-sm font-semibold text-slate-900">Ringkasan PPh 21</h3>
		</div>
		<p class="mt-0.5 text-xs text-slate-500">Metode TER sesuai PMK 168/2023</p>
	</div>

	<div class="p-5">
		<!-- Form input -->
		<div class="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
			<div>
				<label for="pph-salary" class="mb-1.5 block text-xs font-medium text-slate-700">Gaji Bruto</label>
				<div class="relative">
					<span class="absolute left-3 top-1/2 -translate-y-1/2 text-xs text-slate-400">Rp</span>
					<input
						id="pph-salary"
						type="number"
						bind:value={grossSalary}
						min="0"
						step="100000"
						class="w-full rounded-lg border border-slate-200 py-2 pl-9 pr-3 text-sm outline-none focus:border-purple-500 focus:ring-2 focus:ring-purple-100"
					/>
				</div>
			</div>
			<div>
				<label for="pph-marital" class="mb-1.5 block text-xs font-medium text-slate-700">Status Kawin</label>
				<select
					id="pph-marital"
					bind:value={marital}
					class="w-full rounded-lg border border-slate-200 py-2 px-3 text-sm outline-none focus:border-purple-500 focus:ring-2 focus:ring-purple-100"
				>
					<option value="TK">Tidak Kawin (TK)</option>
					<option value="K">Kawin (K)</option>
				</select>
			</div>
			<div>
				<label for="pph-deps" class="mb-1.5 block text-xs font-medium text-slate-700">Tanggungan</label>
				<select
					id="pph-deps"
					bind:value={dependents}
					class="w-full rounded-lg border border-slate-200 py-2 px-3 text-sm outline-none focus:border-purple-500 focus:ring-2 focus:ring-purple-100"
				>
					<option value={0}>0 tanggungan</option>
					<option value={1}>1 tanggungan</option>
					<option value={2}>2 tanggungan</option>
					<option value={3}>3 tanggungan</option>
				</select>
			</div>
		</div>

		<button
			type="button"
			onclick={calculate}
			disabled={loading}
			class="mb-4 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-purple-600 py-2.5 text-sm font-medium text-white transition-all hover:bg-purple-700 disabled:opacity-50"
		>
			{#if loading}
				<span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
			{/if}
			Hitung PPh 21
		</button>

		{#if result}
			<!-- Kategori TER -->
			<div class="mb-4 flex items-start gap-2 rounded-lg bg-purple-50 p-3">
				<Info size={14} class="mt-0.5 shrink-0 text-purple-600" />
				<div>
					<p class="text-xs font-medium text-purple-800">
						{categoryLabel[result.ter_category] ?? result.ter_category}
					</p>
					<p class="text-xs text-purple-600">
						Tarif TER: {result.ter_rate.toFixed(2)}%
					</p>
				</div>
			</div>

			<!-- Hasil kalkulasi -->
			<div class="space-y-2">
				<div class="flex justify-between rounded-lg bg-slate-50 px-4 py-2.5 text-sm">
					<span class="text-slate-600">Gaji Bruto</span>
					<span class="font-medium text-slate-900">{formatCurrency(result.gross_monthly)}</span>
				</div>
				<div class="flex justify-between rounded-lg bg-slate-50 px-4 py-2.5 text-sm">
					<span class="text-slate-600">PTKP (setahun)</span>
					<span class="font-medium text-slate-900">{formatCurrency(result.ptkp)}</span>
				</div>
				<div class="flex justify-between rounded-lg bg-purple-50 px-4 py-2.5 text-sm">
					<span class="font-medium text-purple-700">PPh 21 Bulan Ini</span>
					<span class="font-semibold text-purple-900">{formatCurrency(result.pph21_amount)}</span>
				</div>
				{#if result.is_december}
					<div class="rounded-lg border border-amber-200 bg-amber-50 p-3">
						<p class="mb-1 text-xs font-medium text-amber-800">Koreksi Desember (Tarif Progresif)</p>
						<div class="space-y-1">
							<div class="flex justify-between text-xs">
								<span class="text-amber-700">Penghasilan Bruto Setahun</span>
								<span class="font-medium">{formatCurrency(result.annual_gross)}</span>
							</div>
							<div class="flex justify-between text-xs">
								<span class="text-amber-700">PKP</span>
								<span class="font-medium">{formatCurrency(result.pkp)}</span>
							</div>
							<div class="flex justify-between text-xs">
								<span class="text-amber-700">Pajak Setahun</span>
								<span class="font-medium">{formatCurrency(result.annual_tax)}</span>
							</div>
						</div>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
