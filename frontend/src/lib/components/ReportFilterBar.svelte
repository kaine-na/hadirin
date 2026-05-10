<script lang="ts">
	interface Props {
		startDate: string;
		endDate: string;
		departmentId: string;
		onFilter: (start: string, end: string, dept: string) => void;
	}

	let { startDate = $bindable(), endDate = $bindable(), departmentId = $bindable(), onFilter }: Props = $props();

	// Default: 30 hari terakhir
	const today = new Date();
	const thirtyDaysAgo = new Date(today);
	thirtyDaysAgo.setDate(today.getDate() - 30);

	let localStart = $state(startDate || thirtyDaysAgo.toISOString().split('T')[0]);
	let localEnd = $state(endDate || today.toISOString().split('T')[0]);
	let localDept = $state(departmentId || '');

	function handleFilter() {
		onFilter(localStart, localEnd, localDept);
	}

	function handleReset() {
		localStart = thirtyDaysAgo.toISOString().split('T')[0];
		localEnd = today.toISOString().split('T')[0];
		localDept = '';
		onFilter(localStart, localEnd, localDept);
	}
</script>

<div class="flex flex-wrap items-end gap-3 rounded-xl bg-white p-4 shadow-sm ring-1 ring-slate-200">
	<div class="flex flex-col gap-1">
		<label for="start-date" class="text-xs font-medium text-slate-600">Tanggal Mulai</label>
		<input
			id="start-date"
			type="date"
			bind:value={localStart}
			max={localEnd}
			class="rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100"
		/>
	</div>

	<div class="flex flex-col gap-1">
		<label for="end-date" class="text-xs font-medium text-slate-600">Tanggal Akhir</label>
		<input
			id="end-date"
			type="date"
			bind:value={localEnd}
			min={localStart}
			max={today.toISOString().split('T')[0]}
			class="rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100"
		/>
	</div>

	<div class="flex flex-col gap-1">
		<label for="department" class="text-xs font-medium text-slate-600">Departemen</label>
		<input
			id="department"
			type="text"
			bind:value={localDept}
			placeholder="Semua departemen"
			class="rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100"
		/>
	</div>

	<div class="flex gap-2">
		<button
			type="button"
			onclick={handleFilter}
			class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1"
		>
			Terapkan Filter
		</button>
		<button
			type="button"
			onclick={handleReset}
			class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-300 focus:ring-offset-1"
		>
			Reset
		</button>
	</div>
</div>
