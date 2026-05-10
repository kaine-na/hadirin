<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { documentsApi, downloadBlob } from '$lib/api';
	import { toast } from '$lib/stores/toast.svelte';
	import { auth, isHRorManager } from '$lib/stores/auth.svelte';
	import Button from '$lib/components/Button.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import { formatDate, formatDateTime, formatFileSize } from '$lib/utils/format';
	import type { DocumentItem, DocumentComment } from '$lib/types';

	let doc = $state<DocumentItem | null>(null);
	let comments = $state<DocumentComment[]>([]);
	let loading = $state(true);
	let newComment = $state('');
	let posting = $state(false);

	const id = $derived($page.params.id);

	async function load() {
		loading = true;
		try {
			const [d, c] = await Promise.all([documentsApi.get(id), documentsApi.listComments(id)]);
			doc = d;
			comments = c ?? [];
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal memuat dokumen');
		} finally {
			loading = false;
		}
	}

	async function download() {
		if (!doc) return;
		try {
			const blob = await documentsApi.download(doc.id);
			downloadBlob(blob, doc.file_name);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal mengunduh');
		}
	}

	async function postComment(e: SubmitEvent) {
		e.preventDefault();
		if (!newComment.trim() || !doc) return;
		posting = true;
		try {
			const c = await documentsApi.addComment(doc.id, newComment.trim());
			comments = [...comments, c];
			newComment = '';
			toast.success('Komentar ditambahkan');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Gagal menambah komentar');
		} finally {
			posting = false;
		}
	}

	onMount(load);

	const isImage = $derived(doc?.mime_type?.startsWith('image/') ?? false);
	const isPdf = $derived(doc?.mime_type === 'application/pdf');
</script>

<svelte:head>
	<title>{doc?.title ?? 'Dokumen'} — SaaS Karyawan</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<a href="/documents" class="text-sm text-primary-700 hover:underline">&larr; Kembali ke daftar</a>
	</div>

	{#if loading}
		<div class="card p-10 text-center text-sm text-slate-500">Memuat...</div>
	{:else if !doc}
		<div class="card p-10 text-center text-sm text-slate-500">Dokumen tidak ditemukan.</div>
	{:else}
		<div class="card p-6">
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div>
					<h1 class="text-2xl font-bold text-slate-900">{doc.title}</h1>
					{#if doc.description}
						<p class="mt-2 text-sm text-slate-600">{doc.description}</p>
					{/if}
					<div class="mt-4 flex flex-wrap items-center gap-3 text-xs text-slate-500">
						<Badge color="slate">{doc.category}</Badge>
						<Badge color="primary">v{doc.version}</Badge>
						<span>{doc.file_name}</span>
						<span>·</span>
						<span>{formatFileSize(doc.file_size)}</span>
						<span>·</span>
						<span>Diupload {formatDateTime(doc.created_at)}</span>
						{#if doc.doc_date}
							<span>·</span>
							<span>Tanggal dokumen {formatDate(doc.doc_date)}</span>
						{/if}
					</div>
				</div>
				<Button onclick={download}>Unduh File</Button>
			</div>

			<!-- Preview area -->
			<div class="mt-6 overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
				{#if isImage}
					<img
						src={`/api/documents/${doc.id}/download`}
						alt={doc.title}
						class="mx-auto max-h-[600px] object-contain"
					/>
				{:else if isPdf}
					<iframe
						title={doc.title}
						src={`/api/documents/${doc.id}/download`}
						class="h-[600px] w-full"
					></iframe>
				{:else}
					<div class="p-10 text-center text-sm text-slate-500">
						Preview tidak tersedia untuk tipe file ini. Silakan unduh untuk membuka.
					</div>
				{/if}
			</div>
		</div>

		<!-- Komentar -->
		<div class="card p-6">
			<h2 class="text-lg font-semibold text-slate-900">Komentar</h2>
			<p class="mt-1 text-sm text-slate-500">
				HR dan Manager dapat memberi komentar pada dokumen ini.
			</p>

			<div class="mt-5 space-y-4">
				{#if comments.length === 0}
					<p class="text-sm text-slate-500">Belum ada komentar.</p>
				{:else}
					{#each comments as c}
						<div class="rounded-lg border border-slate-200 bg-slate-50 p-3">
							<div class="flex items-center justify-between text-xs text-slate-500">
								<span class="font-medium text-slate-700">Komentar HR/Manager</span>
								<span>{formatDateTime(c.created_at)}</span>
							</div>
							<p class="mt-2 text-sm text-slate-800 whitespace-pre-wrap">{c.content}</p>
						</div>
					{/each}
				{/if}
			</div>

			{#if isHRorManager()}
				<form class="mt-5 space-y-3" onsubmit={postComment}>
					<div>
						<label for="cm" class="label">Tambah Komentar</label>
						<textarea
							id="cm"
							rows="3"
							class="input"
							placeholder="Tulis komentar untuk dokumen ini..."
							bind:value={newComment}
						></textarea>
					</div>
					<div class="flex justify-end">
						<Button type="submit" loading={posting} disabled={!newComment.trim()}>Kirim</Button>
					</div>
				</form>
			{/if}
		</div>
	{/if}
</div>
