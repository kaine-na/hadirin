<script lang="ts">
	import type { EmployeeTHR, HolidayInfo } from '$lib/api/compliance';
	import { formatCurrency, formatDate } from '$lib/utils/format';
	import { Gift, Calendar, Clock } from 'lucide-svelte';

	interface Props {
		employees: EmployeeTHR[];
		totalTHR: number;
		holidays: HolidayInfo[];
		year: number;
	}

	let { employees, totalTHR, holidays, year }: Props = $props();

	const religionLabel: Record<string, string> = {
		islam: 'Islam',
		kristen: 'Kristen',
		katolik: 'Katolik',
		hindu: 'Hindu',
		buddha: 'Buddha',
		konghucu: 'Konghucu'
	};

	// Ambil hari raya terdekat
	const nearestHoliday = $derived(() => {
		const now = new Date();
		return holidays
			.filter((h) => new Date(h.date) > now)
			.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())[0];
	});

	// Hitung progress bar countdown
	function getCountdownProgress(daysUntilHoliday: number): number {
		const maxDays = 365;
		const remaining = Math.max(0, daysUntilHoliday);
		return Math.max(0, Math.min(100, ((maxDays - remaining) / maxDays) * 100));
	}

	function getProgressColor(daysUntil: number): string {
		if (daysUntil <= 7) return 'bg-red-500';
		if (daysUntil <= 30) return 'bg-amber-500';
		return 'bg-blue-500';
	}
</script>

<div class="rounded-xl border border-slate-200 bg-white shadow-sm">
	<div class="border-b border-slate-100 px-5 py-4">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-2">
				<Gift size={16} class="text-rose-600" strokeWidth={1.75} />
				<h3 class="text-sm font-semibold text-slate-900">Kalkulator THR {year}</h3>
			</div>
			<div class="text-right">
				<p class="text-xs text-slate-500">Total THR</p>
				<p class="text-sm font-semibold text-rose-700">{formatCurrency(totalTHR)}</p>
			</div>
		</div>
	</div>

	<div class="p-5">
		<!-- Countdown hari raya terdekat -->
		{#if nearestHoliday()}
			{@const holiday = nearestHoliday()!}
			<div class="mb-5 rounded-xl bg-gradient-to-r from-rose-50 to-orange-50 p-4">
				<div class="mb-3 flex items-center justify-between">
					<div class="flex items-center gap-2">
						<Calendar size={14} class="text-rose-600" />
						<p class="text-xs font-semibold text-rose-800">{holiday.name}</p>
					</div>
					<p class="text-xs text-rose-600">{formatDate(holiday.date)}</p>
				</div>

				<!-- Progress bar countdown -->
				{#if employees.find((e) => e.religion === holiday.religion)}
					{@const firstEmployee = employees.find((e) => e.religion === holiday.religion)!}
					{@const daysLeft = firstEmployee.thr.days_until_holiday}
					<div class="mb-2">
						<div class="h-2 w-full overflow-hidden rounded-full bg-rose-100">
							<div
								class="h-full rounded-full transition-all {getProgressColor(daysLeft)}"
								style="width: {getCountdownProgress(daysLeft)}%"
							></div>
						</div>
					</div>
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-1.5 text-xs text-rose-700">
							<Clock size={12} />
							<span class="font-medium">{daysLeft} hari</span> lagi
						</div>
						<p class="text-xs text-rose-600">
							Deadline H-7: {firstEmployee.thr.deadline_h7 ? formatDate(firstEmployee.thr.deadline_h7) : '-'}
						</p>
					</div>
				{/if}
			</div>
		{/if}

		<!-- Tabel karyawan -->
		<div class="overflow-x-auto">
			<table class="w-full text-xs">
				<thead>
					<tr class="border-b border-slate-100">
						<th class="pb-2 text-left font-medium text-slate-500">Karyawan</th>
						<th class="pb-2 text-left font-medium text-slate-500">Agama</th>
						<th class="pb-2 text-right font-medium text-slate-500">Masa Kerja</th>
						<th class="pb-2 text-right font-medium text-slate-500">Gaji Pokok</th>
						<th class="pb-2 text-right font-medium text-slate-500">THR</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-50">
					{#each employees as emp (emp.user_id)}
						<tr class="hover:bg-slate-50">
							<td class="py-2.5 font-medium text-slate-900">{emp.name}</td>
							<td class="py-2.5 text-slate-500">{religionLabel[emp.religion] ?? emp.religion}</td>
							<td class="py-2.5 text-right text-slate-600">
								{emp.thr.service_months} bln
							</td>
							<td class="py-2.5 text-right text-slate-600">
								{formatCurrency(emp.thr.base_salary)}
							</td>
							<td class="py-2.5 text-right">
								{#if !emp.thr.is_eligible}
									<span class="text-slate-400">Belum eligible</span>
								{:else}
									<span class="font-semibold text-slate-900">{formatCurrency(emp.thr.thr_amount)}</span>
									{#if !emp.thr.is_full_amount}
										<span class="ml-1 text-amber-600">({(emp.thr.pro_rata_ratio * 100).toFixed(0)}%)</span>
									{/if}
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
				<tfoot>
					<tr class="border-t border-slate-200">
						<td colspan="4" class="pt-3 text-right text-xs font-semibold text-slate-700">Total THR</td>
						<td class="pt-3 text-right text-sm font-bold text-rose-700">{formatCurrency(totalTHR)}</td>
					</tr>
				</tfoot>
			</table>
		</div>

		{#if employees.length === 0}
			<div class="py-8 text-center">
				<p class="text-sm text-slate-500">Belum ada data karyawan</p>
			</div>
		{/if}
	</div>
</div>
