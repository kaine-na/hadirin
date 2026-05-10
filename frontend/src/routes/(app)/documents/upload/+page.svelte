<script lang="ts">
	import { goto } from '$app/navigation';
	import { documentsApi } from '$lib/api';
	import { toast } from '$lib/stores/toast.svelte';
	import Button from '$lib/components/Button.svelte';
	import FileUpload from '$lib/components/FileUpload.svelte';
	import { todayISO } from '$lib/utils/format';

	let title = $state('');
	let description = $state('');
	let category = $state('Laporan Harian');
	let docDate = $state(todayISO());
	let file = $state<File | null>(null);
	let saving = $state(false);

	const categories = ['Laporan Harian', 'Laporan Mingguan', 'Laporan Proyek', 'Lainnya'];

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!file) {
			toast.error('File wajib dipilih');
			return;
		}
		if (!title.trim()) {
			toast.error('Judul wajib diisi');
			return;
		}
		saving = true;
		try {
			await documentsApi.upload({
				title: title.trim(),
				description: description.trim(),
				category,
				doc_date: docDate,
				file
			});
			toast.success('Dokumen berhasil diupload');
			await goto('/documents');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Gagal upload dokumen');
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Upload Dokumen — SaaS Karyawan</title>
</svelte:head>

<div class="mx-auto max-w-2xl space-y-6">
	<div>
		<a href="/documents" class="text-sm text-primary-700 hover:underline">&larr; Kembali ke daftar</a>
		<h1 class="mt-2 text-2xl font-bold text-slate-900">Upload Dokumen</h1>
		<p class="mt-1 text-sm text-slate-500">
			Format yang didukung: PDF, DOCX, XLSX, PNG, JPG. Maksimum 10 MB per file.
		</p>
	</div>

	<form class="card space-y-5 p-6" onsubmit={handleSubmit}>
		<div>
			<label for="title" class="label">Judul <span class="text-red-500">*</span></label>
			<input
				id="title"
				type="text"
				required
				class="input"
				placeholder="Contoh: Laporan Harian 10 Mei 2026"
				bind:value={title}
			/>
		</div>

		<div>
			<label for="desc" class="label">Deskripsi</label>
			<textarea
				id="desc"
				rows="3"
				class="input"
				placeholder="Penjelasan singkat tentang dokumen"
				bind:value={description}
			></textarea>
		</div>

		<div class="grid gap-4 md:grid-cols-2">
			<div>
				<label for="cat" class="label">Kategori <span class="text-red-500">*</span></label>
				<select id="cat" required class="input" bind:value={category}>
					{#each categories as c}
						<option value={c}>{c}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="dt" class="label">Tanggal Dokumen <span class="text-red-500">*</span></label>
				<input id="dt" type="date" required class="input" bind:value={docDate} />
			</div>
		</div>

		<div>
			<p class="label">File <span class="text-red-500">*</span></p>
			<FileUpload
				bind:file
				accept=".pdf,.docx,.xlsx,.png,.jpg,.jpeg,application/pdf,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,image/png,image/jpeg"
				maxSizeMB={10}
				label="Klik untuk pilih"
				helpText="PDF, DOCX, XLSX, PNG, JPG"
			/>
		</div>

		<div class="flex items-center justify-end gap-3 border-t border-slate-200 pt-4">
			<a href="/documents" class="btn-secondary">Batal</a>
			<Button type="submit" loading={saving} disabled={!file || !title.trim()}>Upload Dokumen</Button>
		</div>
	</form>
</div>
