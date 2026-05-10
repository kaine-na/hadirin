<script lang="ts">
	import { onMount } from 'svelte';
	import { documentsApi, employeesApi, downloadBlob } from '$lib/api';
	import { toast } from '$lib/stores/toast.svelte';
	import { auth, isHRorManager } from '$lib/stores/auth.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Table from '$lib/components/Table.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import ConfirmModal from '$lib/components/ui/ConfirmModal.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import type { Column } from '$lib/components/types';
	import { formatDate, formatFileSize } from '$lib/utils/format';
	import type { DocumentItem, Employee } from '$lib/types';
	import { Upload, Download, Trash2, Eye, FileText, Search, Filter, RefreshCw } from 'lucide-svelte';

	interface Row extends DocumentItem {
		ownerName?: string;
	}

	let docs = $state<Row[]>([]);
	let loading = $state(true);
	let employees = $state<Employee[]>([]);

	let filterUser = $state('');
	let filterCategory = $state('');
	let searchQuery = $state('');

	// Pagination state — terhubung ke API params ?page=N&limit=N
	let currentPage = $state(1);
	const PAGE_SIZE = 20;
	let totalDocs = $state(0);

	// Confirm modal state
	let confirmOpen = $state(false);
	let confirmDoc = $state<Row | null>(null);

	const categories = ['Laporan Harian', 'Laporan Mingguan', 'Laporan Proyek', 'Lainnya'];

	const columns: Column<Row>[] = [
		{ key: 'title', label: 'Judul' },
		{ key: 'category', label: 'Kategori' },
		{ key: 'owner', label: 'Pemilik' },
		{ key: 'size', label: 'Ukuran' },
		{ key: 'version', label: 'Versi' },
		{ key: 'created_at', label: 'Tanggal' },
		{ key: 'actions', label: '', align: 'right' }
	];

	async function loadEmployees() {
		if (!isHRorManager()) return;
		try {
			const res = await employeesApi.list({ page: 1, page_size: 500 });
			employees = res.items;
		} catch {
			// abaikan
		}
	}

	async function loadDocs() {
		loading = true;
		try {
			const res = await documentsApi.list({
				user_id: filterUser || undefined,
				category: filterCategory || undefined,
				page: currentPage,
				page_size: PAGE_SIZE
			});
			const empMap = new Map(employees.map((e) => [e.id, e.name]));
			docs = res.items.map((d) => ({
				...d,
				ownerName: empMap.get(d.user_id) ?? (d.user_id === auth.user?.id ? auth.user?.name : '-')
			}));
			totalDocs = res.total ?? res.items.length;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal memuat dokumen');
		} finally {
			loading = false;
		}
	}

	function handlePageChange(page: number) {
		currentPage = page;
		loadDocs();
	}

	function handleFilterApply() {
		currentPage = 1;
		loadDocs();
	}

	async function download(doc: DocumentItem) {
		try {
			const blob = await documentsApi.download(doc.id);
			downloadBlob(blob, doc.file_name);
			toast.success('Dokumen berhasil diunduh');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal mengunduh');
		}
	}

	function confirmRemove(doc: Row) {
		confirmDoc = doc;
		confirmOpen = true;
	}

	async function doRemove() {
		if (!confirmDoc) return;
		try {
			await documentsApi.remove(confirmDoc.id);
			toast.success('Dokumen berhasil dihapus');
			await loadDocs();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal menghapus');
		} finally {
			confirmOpen = false;
			confirmDoc = null;
		}
	}

	onMount(async () => {
		await loadEmployees();
		await loadDocs();
	});

	const filteredDocs = $derived(
		searchQuery
			? docs.filter(
					(d) =>
						d.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
						d.ownerName?.toLowerCase().includes(searchQuery.toLowerCase())
				)
			: docs
	);
</script>

<svelte:head>
	<title>Dokumen — Hadir</title>
</svelte:head>

<ConfirmModal
	open={confirmOpen}
	title="Hapus Dokumen"
	message={`Apakah Anda yakin ingin menghapus dokumen "${confirmDoc?.title}"? Tindakan ini tidak dapat dibatalkan.`}
	confirmLabel="Hapus"
	danger={true}
	onConfirm={doRemove}
	onCancel={() => {
		confirmOpen = false;
		confirmDoc = null;
	}}
/>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<PageHeader
			title="Berkas Kerja"
			description="Upload dan kelola laporan kerja Anda. Semua dokumen tersimpan aman."
		/>
		<a
			href="/documents/upload"
			class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition-all duration-150 hover:bg-blue-700 hover:shadow-md active:scale-95"
		>
			<Upload size={16} strokeWidth={2} />
			Upload Dokumen
		</a>
	</div>

	<!-- Filter Bar -->
	<div class="rounded-xl border border-slate-100 bg-white p-5 shadow-sm">
		<div class="grid gap-4 md:grid-cols-4">
			<!-- Search -->
			<div class="relative md:col-span-2">
				<Search
					size={16}
					strokeWidth={1.75}
					class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
				/>
				<input
					type="text"
					placeholder="Cari dokumen atau nama karyawan..."
					class="w-full rounded-lg border border-slate-200 bg-white py-2 pl-9 pr-4 text-sm text-slate-900 placeholder-slate-400 shadow-sm transition-all focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
					bind:value={searchQuery}
				/>
			</div>

			{#if isHRorManager()}
				<div>
					<select
						class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 shadow-sm transition-all focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
						bind:value={filterUser}
					>
						<option value="">Semua karyawan</option>
						{#each employees as emp}
							<option value={emp.id}>{emp.name}</option>
						{/each}
					</select>
				</div>
			{/if}

			<div>
				<select
					class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 shadow-sm transition-all focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
					bind:value={filterCategory}
				>
					<option value="">Semua kategori</option>
					{#each categories as c}
						<option value={c}>{c}</option>
					{/each}
				</select>
			</div>

			<div class="flex items-center gap-2">
				<button
					type="button"
					onclick={handleFilterApply}
					disabled={loading}
					class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm transition-all duration-150 hover:bg-slate-50 hover:shadow-md active:scale-95 disabled:opacity-50"
				>
					{#if loading}
						<RefreshCw size={14} strokeWidth={2} class="animate-spin" />
					{:else}
						<Filter size={14} strokeWidth={2} />
					{/if}
					Terapkan
				</button>
			</div>
		</div>
	</div>

	<!-- Tabel Dokumen -->
	{#if !loading && filteredDocs.length === 0}
		<EmptyState
			icon={FileText}
			title="Belum ada dokumen"
			description="Upload dokumen laporan kerja Anda untuk memulai."
			ctaLabel="Upload Dokumen"
			ctaHref="/documents/upload"
		/>
	{:else}
		<div class="rounded-xl border border-slate-100 bg-white shadow-sm overflow-hidden">
			<Table {columns} rows={filteredDocs} loading={loading} emptyMessage="Belum ada dokumen">
				{#snippet row(d: Row)}
					<tr class="group transition-colors hover:bg-slate-50">
						<td class="px-4 py-3">
							<div class="flex items-center gap-3">
								<div
									class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600"
								>
									<FileText size={14} strokeWidth={1.75} />
								</div>
								<div class="min-w-0">
									<a
										href={`/documents/${d.id}`}
										class="text-sm font-medium text-blue-700 hover:text-blue-800 hover:underline"
									>
										{d.title}
									</a>
									{#if d.description}
										<p class="mt-0.5 line-clamp-1 text-xs text-slate-500">{d.description}</p>
									{/if}
								</div>
							</div>
						</td>
						<td class="px-4 py-3 text-sm">
							<Badge color="slate">{d.category}</Badge>
						</td>
						<td class="px-4 py-3 text-sm text-slate-700">{d.ownerName ?? '-'}</td>
						<td class="px-4 py-3 text-sm text-slate-500">{formatFileSize(d.file_size)}</td>
						<td class="px-4 py-3 text-sm text-slate-500">v{d.version}</td>
						<td class="px-4 py-3 text-sm text-slate-500">{formatDate(d.created_at)}</td>
						<td class="px-4 py-3 text-right text-sm">
							<div class="flex justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100">
								<a
									href={`/documents/${d.id}`}
									class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700"
									title="Lihat detail"
								>
									<Eye size={14} strokeWidth={1.75} />
								</a>
								<button
									type="button"
									class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-blue-50 hover:text-blue-600"
									onclick={() => download(d)}
									title="Unduh"
								>
									<Download size={14} strokeWidth={1.75} />
								</button>
								{#if d.user_id === auth.user?.id || isHRorManager()}
									<button
										type="button"
										class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-red-50 hover:text-red-600"
										onclick={() => confirmRemove(d)}
										title="Hapus"
									>
										<Trash2 size={14} strokeWidth={1.75} />
									</button>
								{/if}
							</div>
						</td>
					</tr>
				{/snippet}
			</Table>
			<!-- Pagination terhubung ke API params -->
			<Pagination
				page={currentPage}
				pageSize={PAGE_SIZE}
				total={totalDocs}
				onPageChange={handlePageChange}
			/>
		</div>
	{/if}
</div>
