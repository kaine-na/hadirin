<script lang="ts">
	import { formatFileSize } from '$lib/utils/format';

	interface Props {
		accept?: string;
		maxSizeMB?: number;
		label?: string;
		helpText?: string;
		file?: File | null;
		onchange?: (file: File | null) => void;
	}

	let {
		accept = '',
		maxSizeMB = 10,
		label = 'Pilih file',
		helpText = '',
		file = $bindable(null),
		onchange
	}: Props = $props();

	let dragging = $state(false);
	let error = $state('');
	let inputEl: HTMLInputElement | undefined = $state();

	function handleFile(f: File | null | undefined) {
		error = '';
		if (!f) {
			file = null;
			onchange?.(null);
			return;
		}
		const maxBytes = maxSizeMB * 1024 * 1024;
		if (f.size > maxBytes) {
			error = `Ukuran file melebihi batas ${maxSizeMB} MB`;
			file = null;
			onchange?.(null);
			return;
		}
		file = f;
		onchange?.(f);
	}

	function onInput(e: Event) {
		const target = e.currentTarget as HTMLInputElement;
		handleFile(target.files?.[0] ?? null);
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		handleFile(e.dataTransfer?.files?.[0] ?? null);
	}

	function openPicker() {
		inputEl?.click();
	}

	function remove() {
		handleFile(null);
		if (inputEl) inputEl.value = '';
	}
</script>

<div>
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="flex flex-col items-center justify-center rounded-lg border-2 border-dashed px-6 py-8 text-center transition-colors {dragging
			? 'border-primary-500 bg-primary-50'
			: 'border-slate-300 bg-slate-50'}"
		ondragover={(e) => {
			e.preventDefault();
			dragging = true;
		}}
		ondragleave={() => (dragging = false)}
		ondrop={onDrop}
	>
		<svg class="mb-2 h-10 w-10 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="1.5"
				d="M7 16V8a4 4 0 118 0v8m-4-12v12m0 0l-4-4m4 4l4-4M4 20h16"
			/>
		</svg>
		{#if file}
			<p class="text-sm font-medium text-slate-900">{file.name}</p>
			<p class="mt-1 text-xs text-slate-500">{formatFileSize(file.size)}</p>
			<div class="mt-3 flex gap-2">
				<button
					type="button"
					class="text-sm font-medium text-primary-700 hover:text-primary-800"
					onclick={openPicker}
				>
					Ganti file
				</button>
				<span class="text-slate-300">•</span>
				<button
					type="button"
					class="text-sm font-medium text-red-600 hover:text-red-700"
					onclick={remove}
				>
					Hapus
				</button>
			</div>
		{:else}
			<p class="text-sm text-slate-600">
				<button
					type="button"
					class="font-medium text-primary-700 hover:text-primary-800"
					onclick={openPicker}
				>
					{label}
				</button>
				atau drag & drop
			</p>
			{#if helpText}
				<p class="mt-1 text-xs text-slate-500">{helpText}</p>
			{/if}
			<p class="mt-1 text-xs text-slate-500">Maks. {maxSizeMB} MB</p>
		{/if}
		<input
			bind:this={inputEl}
			type="file"
			class="hidden"
			{accept}
			onchange={onInput}
		/>
	</div>
	{#if error}
		<p class="mt-2 text-sm text-red-600">{error}</p>
	{/if}
</div>
