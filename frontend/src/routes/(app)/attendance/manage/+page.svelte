<script lang="ts">
	import { onMount } from 'svelte';
	import { attendanceApi, employeesApi, downloadBlob } from '$lib/api';
	import { toast } from '$lib/stores/toast.svelte';
	import { isHRorManager, isHR } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Table from '$lib/components/Table.svelte';
	import type { Column } from '$lib/components/types';
	import Modal from '$lib/components/Modal.svelte';
	import { formatDate, formatTime, statusLabel, statusColor, startOfMonthISO, todayISO } from '$lib/utils/format';
	import type { Attendance, AttendanceStatus, Employee } from '$lib/types';

	interface AttendanceRow extends Attendance {
		employeeName?: string;
	}

	let employees = $state<Employee[]>([]);
	let selectedEmployeeId = $state<string>('');
	let startDate = $state(startOfMonthISO());
	let endDate = $state(todayISO());

	let records = $state<AttendanceRow[]>([]);
	let loading = $state(false);
	let exporting = $state(false);

	// Modal override
	let editOpen = $state(false);
	let editRow = $state<Attendance | null>(null);
	let editStatus = $state<AttendanceStatus>('hadir');
	let editNotes = $state('');
	let editClockIn = $state('');
	let editClockOut = $state('');
	let saving = $state(false);

	const columns: Column<AttendanceRow>[] = [
		{ key: 'employeeName', label: 'Karyawan' },
		{ key: 'date', label: 'Tanggal' },
		{ key: 'clock_in', label: 'Clock In' },
		{ key: 'clock_out', label: 'Clock Out' },
		{ key: 'status', label: 'Status' },
		{ key: 'actions', label: 'Aksi', align: 'right' }
	];

	onMount(async () => {
		if (!isHRorManager()) {
			goto('/dashboard');
			return;
		}
		try {
			const res = await employeesApi.list({ page: 1, page_size: 500 });
			employees = res.items;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal memuat karyawan');
		}
	});

	async function loadData() {
		loading = true;
		try {
			if (selectedEmployeeId) {
				const res = await attendanceApi.byEmployee(selectedEmployeeId, {
					start_date: startDate,
					end_date: endDate,
					page: 1,
					page_size: 200
				});
				const emp = employees.find((e) => e.id === selectedEmployeeId);
				records = res.items.map((r) => ({ ...r, employeeName: emp?.name ?? '-' }));
			} else {
				// Gabungkan semua karyawan
				const all = await Promise.all(
					employees.map(async (emp) => {
						const res = await attendanceApi.byEmployee(emp.id, {
							start_date: startDate,
							end_date: endDate,
							page: 1,
							page_size: 200
						});
						return res.items.map((r) => ({ ...r, employeeName: emp.name }));
					})
				);
				records = all.flat().sort((a, b) => (a.date < b.date ? 1 : -1));
			}
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal memuat rekap');
		} finally {
			loading = false;
		}
	}

	async function exportCsv() {
		exporting = true;
		try {
			const blob = await attendanceApi.exportCsv({ start_date: startDate, end_date: endDate });
			downloadBlob(blob, `absensi_${startDate}_${endDate}.csv`);
			toast.success('File CSV berhasil diunduh');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal export CSV');
		} finally {
			exporting = false;
		}
	}

	function openEdit(row: Attendance) {
		editRow = row;
		editStatus = row.status;
		editNotes = row.notes ?? '';
		editClockIn = row.clock_in ?? '';
		editClockOut = row.clock_out ?? '';
		editOpen = true;
	}

	async function saveOverride() {
		if (!editRow) return;
		saving = true;
		try {
			await attendanceApi.override(editRow.id, {
				status: editStatus,
				notes: editNotes,
				clock_in: editClockIn,
				clock_out: editClockOut
			});
			toast.success('Data absensi berhasil diperbarui');
			editOpen = false;
			editRow = null;
			await loadData();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal menyimpan');
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Kelola Absensi — SaaS Karyawan</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-slate-900">Kelola Absensi</h1>
			<p class="mt-1 text-sm text-slate-500">Lihat dan koreksi catatan absensi seluruh karyawan.</p>
		</div>
		<Button variant="secondary" onclick={exportCsv} loading={exporting}>
			<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
			</svg>
			Export CSV
		</Button>
	</div>

	<!-- Filter -->
	<div class="card p-5">
		<div class="grid gap-4 md:grid-cols-4">
			<div>
				<label for="emp" class="label">Karyawan</label>
				<select id="emp" class="input" bind:value={selectedEmployeeId}>
					<option value="">Semua karyawan</option>
					{#each employees as emp}
						<option value={emp.id}>{emp.name}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="start" class="label">Dari</label>
				<input id="start" type="date" class="input" bind:value={startDate} />
			</div>
			<div>
				<label for="end" class="label">Sampai</label>
				<input id="end" type="date" class="input" bind:value={endDate} />
			</div>
			<div class="flex items-end">
				<Button fullWidth onclick={loadData} loading={loading}>Tampilkan</Button>
			</div>
		</div>
	</div>

	<Table {columns} rows={records} loading={loading} emptyMessage="Pilih filter untuk menampilkan data">
		{#snippet row(r: AttendanceRow)}
			<tr class="hover:bg-slate-50">
				<td class="px-4 py-3 text-sm font-medium text-slate-900">{r.employeeName ?? '-'}</td>
				<td class="px-4 py-3 text-sm text-slate-700">{formatDate(r.date)}</td>
				<td class="px-4 py-3 text-sm text-slate-700">{formatTime(r.clock_in)}</td>
				<td class="px-4 py-3 text-sm text-slate-700">{formatTime(r.clock_out)}</td>
				<td class="px-4 py-3 text-sm">
					<Badge color={statusColor(r.status)}>{statusLabel(r.status)}</Badge>
				</td>
				<td class="px-4 py-3 text-right text-sm">
					{#if isHR()}
						<button
							type="button"
							class="font-medium text-primary-700 hover:text-primary-800"
							onclick={() => openEdit(r)}
						>
							Koreksi
						</button>
					{/if}
				</td>
			</tr>
		{/snippet}
	</Table>
</div>

<Modal bind:open={editOpen} title="Koreksi Absensi" size="md">
	<div class="space-y-4">
		<div>
			<label for="st" class="label">Status</label>
			<select id="st" class="input" bind:value={editStatus}>
				<option value="hadir">Hadir</option>
				<option value="terlambat">Terlambat</option>
				<option value="izin">Izin</option>
				<option value="sakit">Sakit</option>
				<option value="alpha">Alpha</option>
			</select>
		</div>
		<div class="grid gap-4 md:grid-cols-2">
			<div>
				<label for="ci" class="label">Clock In (ISO 8601)</label>
				<input
					id="ci"
					type="datetime-local"
					class="input"
					value={editClockIn ? editClockIn.slice(0, 16) : ''}
					oninput={(e) => {
						const v = (e.currentTarget as HTMLInputElement).value;
						editClockIn = v ? new Date(v).toISOString() : '';
					}}
				/>
			</div>
			<div>
				<label for="co" class="label">Clock Out</label>
				<input
					id="co"
					type="datetime-local"
					class="input"
					value={editClockOut ? editClockOut.slice(0, 16) : ''}
					oninput={(e) => {
						const v = (e.currentTarget as HTMLInputElement).value;
						editClockOut = v ? new Date(v).toISOString() : '';
					}}
				/>
			</div>
		</div>
		<div>
			<label for="nt" class="label">Keterangan</label>
			<textarea id="nt" rows="3" class="input" bind:value={editNotes}></textarea>
		</div>
	</div>
	{#snippet footer()}
		<Button variant="secondary" onclick={() => (editOpen = false)}>Batal</Button>
		<Button onclick={saveOverride} loading={saving}>Simpan</Button>
	{/snippet}
</Modal>
