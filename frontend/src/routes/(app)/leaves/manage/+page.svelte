<script lang="ts">
	import { onMount } from 'svelte';
	import { leavesApi, type LeaveRequest } from '$lib/api/leaves';
	import { toast } from '$lib/stores/toast.svelte';
	import SkeletonLoader from '$lib/components/ui/SkeletonLoader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import LeaveApprovalCard from '$lib/components/LeaveApprovalCard.svelte';
	import { CalendarOff, Filter } from 'lucide-svelte';

	let requests = $state<LeaveRequest[]>([]);
	let total = $state(0);
	let page = $state(1);
	const pageSize = 12;

	let loading = $state(true);
	let filterStatus = $state('pending');

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
			toast.error('Gagal memuat daftar pengajuan cuti');
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadRequests();
	});

	$effect(() => {
		loadRequests();
	});
</script>

<svelte:head>
	<title>Kelola Cuti — Hadir</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="mb-6 flex items-center gap-3">
		<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-teal-100">
			<CalendarOff class="h-5 w-5 text-teal-600" />
		</div>
		<div>
			<h1 class="text-2xl font-bold leading-tight text-slate-800">Kelola Cuti</h1>
			<p class="mt-1 text-sm text-slate-500">Review dan proses pengajuan cuti karyawan</p>
		</div>
	</div>

	<!-- Filter Status -->
	<div class="flex items-center gap-3">
		<Filter size={16} strokeWidth={1.75} class="text-slate-400" />
		<div class="flex flex-wrap gap-2">
			{#each [
				{ value: 'pending', label: 'Menunggu Approval' },
				{ value: '', label: 'Semua' },
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
						? 'bg-teal-600 text-white'
						: 'bg-slate-100 text-slate-600 hover:bg-slate-200'}"
				>
					{f.label}
				</button>
			{/each}
		</div>
	</div>

	<!-- Daftar Pengajuan -->
	{#if loading}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array(6) as _}
				<SkeletonLoader rows={4} />
			{/each}
		</div>
	{:else if requests.length === 0}
		<EmptyState
			title={filterStatus === 'pending' ? 'Tidak ada pengajuan yang menunggu' : 'Tidak ada data cuti'}
			description={filterStatus === 'pending'
				? 'Semua pengajuan cuti sudah diproses'
				: 'Belum ada pengajuan cuti dengan filter ini'}
			icon={CalendarOff}
		/>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each requests as req (req.id)}
				<LeaveApprovalCard request={req} onUpdate={loadRequests} />
			{/each}
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
