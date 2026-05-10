<script lang="ts">
	import { leavesApi, type LeaveType } from '$lib/api/leaves';
	import { toast } from '$lib/stores/toast.svelte';

	interface Props {
		onSuccess?: () => void;
		onCancel?: () => void;
	}

	let { onSuccess, onCancel }: Props = $props();

	let leaveTypes = $state<LeaveType[]>([]);
	let loading = $state(false);
	let submitting = $state(false);

	let form = $state({
		leave_type_id: '',
		start_date: '',
		end_date: '',
		reason: ''
	});

	// Hitung total hari otomatis
	const totalDays = $derived(() => {
		if (!form.start_date || !form.end_date) return 0;
		const start = new Date(form.start_date);
		const end = new Date(form.end_date);
		if (end < start) return 0;
		return Math.floor((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)) + 1;
	});

	// Load jenis cuti saat komponen mount
	$effect(() => {
		loadLeaveTypes();
	});

	async function loadLeaveTypes() {
		loading = true;
		try {
			leaveTypes = await leavesApi.getTypes();
		} catch {
			toast.error('Gagal memuat jenis cuti');
		} finally {
			loading = false;
		}
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!form.leave_type_id) {
			toast.error('Pilih jenis cuti terlebih dahulu');
			return;
		}
		if (!form.start_date || !form.end_date) {
			toast.error('Tanggal mulai dan selesai wajib diisi');
			return;
		}
		if (new Date(form.end_date) < new Date(form.start_date)) {
			toast.error('Tanggal selesai tidak boleh sebelum tanggal mulai');
			return;
		}
		if (!form.reason.trim()) {
			toast.error('Alasan cuti wajib diisi');
			return;
		}

		submitting = true;
		try {
			await leavesApi.create(form);
			toast.success('Pengajuan cuti berhasil dikirim');
			onSuccess?.();
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : 'Gagal mengajukan cuti';
			toast.error(msg);
		} finally {
			submitting = false;
		}
	}
</script>

<form onsubmit={handleSubmit} class="space-y-4">
	<!-- Jenis Cuti -->
	<div>
		<label for="leave_type" class="block text-sm font-medium text-slate-700">
			Jenis Cuti <span class="text-red-500">*</span>
		</label>
		<select
			id="leave_type"
			bind:value={form.leave_type_id}
			disabled={loading || submitting}
			class="mt-1 block w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-slate-50 disabled:text-slate-400"
		>
			<option value="">-- Pilih jenis cuti --</option>
			{#each leaveTypes as lt}
				<option value={lt.id}>{lt.name} (maks. {lt.max_days} hari)</option>
			{/each}
		</select>
	</div>

	<!-- Tanggal -->
	<div class="grid grid-cols-2 gap-3">
		<div>
			<label for="start_date" class="block text-sm font-medium text-slate-700">
				Tanggal Mulai <span class="text-red-500">*</span>
			</label>
			<input
				id="start_date"
				type="date"
				bind:value={form.start_date}
				disabled={submitting}
				class="mt-1 block w-full rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-slate-50"
			/>
		</div>
		<div>
			<label for="end_date" class="block text-sm font-medium text-slate-700">
				Tanggal Selesai <span class="text-red-500">*</span>
			</label>
			<input
				id="end_date"
				type="date"
				bind:value={form.end_date}
				min={form.start_date}
				disabled={submitting}
				class="mt-1 block w-full rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-slate-50"
			/>
		</div>
	</div>

	<!-- Total hari -->
	{#if totalDays() > 0}
		<p class="text-sm text-slate-600">
			Total: <span class="font-semibold text-blue-600">{totalDays()} hari</span>
		</p>
	{/if}

	<!-- Alasan -->
	<div>
		<label for="reason" class="block text-sm font-medium text-slate-700">
			Alasan Cuti <span class="text-red-500">*</span>
		</label>
		<textarea
			id="reason"
			bind:value={form.reason}
			disabled={submitting}
			rows={3}
			placeholder="Jelaskan alasan pengajuan cuti..."
			class="mt-1 block w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-slate-50"
		></textarea>
	</div>

	<!-- Actions -->
	<div class="flex justify-end gap-3 pt-2">
		{#if onCancel}
			<button
				type="button"
				onclick={onCancel}
				disabled={submitting}
				class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:opacity-50"
			>
				Batal
			</button>
		{/if}
		<button
			type="submit"
			disabled={submitting || loading}
			class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-all hover:bg-blue-700 hover:shadow-md active:scale-95 disabled:opacity-50"
		>
			{#if submitting}
				<span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"
				></span>
				Mengirim...
			{:else}
				Ajukan Cuti
			{/if}
		</button>
	</div>
</form>
