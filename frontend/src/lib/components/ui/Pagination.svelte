<script lang="ts">
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';

	interface Props {
		page: number;
		pageSize: number;
		total: number;
		onPageChange: (page: number) => void;
	}

	let { page, pageSize, total, onPageChange }: Props = $props();

	const totalPages = $derived(Math.ceil(total / pageSize));
	const from = $derived(total === 0 ? 0 : (page - 1) * pageSize + 1);
	const to = $derived(Math.min(page * pageSize, total));

	// Buat array halaman yang ditampilkan (max 5 halaman di sekitar halaman aktif)
	const pages = $derived(() => {
		if (totalPages <= 7) {
			return Array.from({ length: totalPages }, (_, i) => i + 1);
		}
		const result: (number | '...')[] = [];
		result.push(1);
		if (page > 3) result.push('...');
		for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) {
			result.push(i);
		}
		if (page < totalPages - 2) result.push('...');
		result.push(totalPages);
		return result;
	});
</script>

{#if totalPages > 1}
	<div class="flex items-center justify-between border-t border-slate-200 px-4 py-3">
		<!-- Info -->
		<p class="text-sm text-slate-500">
			Menampilkan <span class="font-medium text-slate-700">{from}–{to}</span> dari
			<span class="font-medium text-slate-700">{total}</span> data
		</p>

		<!-- Navigasi -->
		<div class="flex items-center gap-1">
			<!-- Prev -->
			<button
				type="button"
				onclick={() => onPageChange(page - 1)}
				disabled={page <= 1}
				class="flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 text-slate-500 transition-colors
					hover:bg-slate-50 hover:text-slate-700 disabled:cursor-not-allowed disabled:opacity-40"
				aria-label="Halaman sebelumnya"
			>
				<ChevronLeft size={16} />
			</button>

			<!-- Nomor halaman -->
			{#each pages() as p}
				{#if p === '...'}
					<span class="flex h-8 w-8 items-center justify-center text-sm text-slate-400">…</span>
				{:else}
					<button
						type="button"
						onclick={() => onPageChange(p as number)}
						class="flex h-8 w-8 items-center justify-center rounded-md border text-sm font-medium transition-colors
							{p === page
							? 'border-blue-600 bg-blue-600 text-white'
							: 'border-slate-200 text-slate-600 hover:bg-slate-50 hover:text-slate-900'}"
					>
						{p}
					</button>
				{/if}
			{/each}

			<!-- Next -->
			<button
				type="button"
				onclick={() => onPageChange(page + 1)}
				disabled={page >= totalPages}
				class="flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 text-slate-500 transition-colors
					hover:bg-slate-50 hover:text-slate-700 disabled:cursor-not-allowed disabled:opacity-40"
				aria-label="Halaman berikutnya"
			>
				<ChevronRight size={16} />
			</button>
		</div>
	</div>
{/if}
