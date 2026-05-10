<script lang="ts">
	import { onMount } from 'svelte';
	import { leavesApi, type LeaveRequest, type LeaveBalance } from '$lib/api/leaves';
	import { auth } from '$lib/stores/auth.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import SkeletonLoader from '$lib/components/ui/SkeletonLoader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import LeaveStatusBadge from '$lib/components/LeaveStatusBadge.svelte';
	import LeaveRequestForm from '$lib/components/LeaveRequestForm.svelte';
	import LeaveBalanceWidget from '$lib/components/LeaveBalanceWidget.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import { CalendarOff, Plus, Filter } from 'lucide-svelte';

	let requests = $state<LeaveRequest[]>([]);
	let balances = $state<LeaveBalance[]>([]);
	let total = $state(0);
	let page = $state(1);
	const pageSize = 10;

	let loading = $state(true);
	let loadingBalance = $state(true);
	let showForm = $state(false);

	let filterStatus = $state('');

	async function loadRequests() {
		loading = true;
		try {
			const result = await leavesApi.list({
				status: filterStatus || undefined,
				page,
				page_size: pageSize
			});
			requests = result.items ?? [];
			total = result.total;
		} catch {
			toast.error('Gagal memuat daftar cuti');
		} finally {
			loading = false;
		}
	}

	async function loadBalance() {
		loadingBalance = true;
		try {
			balances = await leavesApi.getMyBalance();
		} catch {
			// abaikan
		} finally {
			loadingBalance = false;
		}
	}

	async function handleCancel(id: string) {
		try {
			await leavesApi.cancel(id);
			toast.success('Pengajuan cuti berhasil dibatalkan');
			await loadRequests();
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : 'Gagal membatalkan cuti';
			toast.error(msg);
		}
	}

	function onFormSuccess() {
		showForm = false;
		loadRequests();
		loadBalance();
	}

	onMount(() => {
		loadRequests();
		loadBalance();
	});

	$effect(() => {
		// Re-load saat filter atau page berubah
		loadRequests();
	});
</script>

<svelte:head>
	<title>Cuti — Hadir</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="mb-6 flex items-center justify-between gap-4">
		<div class="flex items-center gap-3">
			<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-blue-100">
				<CalendarOff class="h-5 w-5 text-blue-600" />
			</div>
			<div>
				<h1 class="text-2xl font-bold leading-tight text-slate-800">Cuti Saya</h1>
				<p class="mt-1 text-sm text-slate-500">Ajukan dan pantau status pengajuan cuti kamu</p>
			</div>
		</div>
		<button
			onclick={() => (showForm = true)}
			class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-all hover:bg-blue-700 hover:shadow-md active:scale-95"
		>
			<Plus size={16} strokeWidth={2} />
			Ajukan Cuti
		</button>
	</div>

	<!-- Saldo Cuti Widget -->
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		<LeaveBalanceWidget {balances} loading={loadingBalance} />

		{#if !loadingBalance && balances.length > 1}
			{#each balances.slice(1, 3) as balance}
				<div
					class="rounded-xl border border-slate-100 bg-white p-5 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:shadow-md"
				>
					<p class="text-sm font-medium text-slate-500">{balance.leave_type_name}</p>
					<p class="mt-1 text-2xl font-bold text-slate-900">
						{balance.remaining_days}
						<span class="text-sm font-normal text-slate-500">/ {balance.total_days} hari</span>
					</p>
					<div class="mt-2 h-1.5 overflow-hidden rounded-full bg-slate-100">
						<div
							class="h-full rounded-full bg-blue-500 transition-all duration-500"
							style="width: {balance.total_days > 0
								? Math.min(100, (balance.used_days / balance.total_days) * 100)
								: 0}%"
						></div>
					</div>
				</div>
			{/each}
		{/if}
	</div>

	<!-- Filter -->
	<div class="flex items-center gap-3">
		<Filter size={16} strokeWidth={1.75} class="text-slate-400" />
		<div class="flex gap-2">
			{#each [
				{ value: '', label: 'Semua' },
				{ value: 'pending', label: 'Menunggu' },
				{ value: 'approved', label: 'Disetujui' },
				{ value: 'rejected', label: 'Ditolak' },
				{ value: 'cancelled', label: 'Dibatalkan' }
			] as f}
				<button
					onclick={() => {
						filterStatus = f.value;
						page = 1;
					}}
					class="rounded-full px-3 py-1 text-xs font-medium transition-colors
					{filterStatus === f.value
						? 'bg-blue-600 text-white'
						: 'bg-slate-100 text-slate-600 hover:bg-slate-200'}"
				>
					{f.label}
				</button>
			{/each}
		</div>
	</div>

	<!-- Daftar Pengajuan -->
	{#if loading}
		<SkeletonLoader rows={5} />
	{:else if requests.length === 0}
		<EmptyState
			title="Belum ada pengajuan cuti"
			description="Klik tombol 'Ajukan Cuti' untuk membuat pengajuan baru"
			icon={CalendarOff}
		/>
	{:else}
		<div class="overflow-hidden rounded-xl border border-slate-100 bg-white shadow-sm">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-slate-100 bg-slate-50">
						<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">Jenis</th>
						<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">Tanggal</th>
						<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">Durasi</th>
						<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">Alasan</th>
						<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">Status</th>
						<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">Aksi</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-50">
					{#each requests as req}
						<tr class="transition-colors hover:bg-slate-50">
							<td class="px-4 py-3 font-medium text-slate-900">{req.leave_type_name ?? '-'}</td>
							<td class="px-4 py-3 text-slate-600">
								{new Date(req.start_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}
								{#if req.start_date !== req.end_date}
									<span class="text-slate-400"> – </span>
									{new Date(req.end_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}
								{/if}
							</td>
							<td class="px-4 py-3 text-slate-600">{req.total_days} hari</td>
							<td class="max-w-xs px-4 py-3 text-slate-600">
								<p class="truncate">{req.reason}</p>
							</td>
							<td class="px-4 py-3">
								<LeaveStatusBadge status={req.status} size="sm" />
							</td>
							<td class="px-4 py-3">
								{#if req.status === 'pending' && req.user_id === auth.user?.id}
									<button
										onclick={() => handleCancel(req.id)}
										class="text-xs font-medium text-red-600 hover:text-red-700"
									>
										Batalkan
									</button>
								{:else if req.status === 'rejected' && req.rejection_reason}
									<span class="text-xs text-slate-400" title={req.rejection_reason}>Lihat alasan</span>
								{:else}
									<span class="text-xs text-slate-300">—</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		{#if total > pageSize}
			<Pagination
				{total}
				{page}
				{pageSize}
				onPageChange={(p) => {
					page = p;
				}}
			/>
		{/if}
	{/if}
</div>

<!-- Modal Form Pengajuan -->
<Modal bind:open={showForm} title="Ajukan Cuti">
	<LeaveRequestForm onSuccess={onFormSuccess} onCancel={() => (showForm = false)} />
</Modal>
