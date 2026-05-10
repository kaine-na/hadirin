<script lang="ts" generics="T">
	import type { Column } from './types';

	interface Props {
		columns: Column<T>[];
		rows: T[];
		loading?: boolean;
		emptyMessage?: string;
		row?: import('svelte').Snippet<[T, number]>;
	}

	let { columns, rows, loading = false, emptyMessage = 'Belum ada data', row }: Props = $props();
</script>

<div class="card overflow-hidden">
	<div class="overflow-x-auto">
		<table class="min-w-full divide-y divide-slate-200">
			<thead class="bg-slate-50">
				<tr>
					{#each columns as col}
						<th
							scope="col"
							class="px-4 py-3 text-{col.align ?? 'left'} text-xs font-semibold uppercase tracking-wider text-slate-600"
							style={col.width ? `width: ${col.width};` : undefined}
						>
							{col.label}
						</th>
					{/each}
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-200 bg-white">
				{#if loading}
					<tr>
						<td colspan={columns.length} class="px-4 py-8 text-center text-sm text-slate-500">
							<div class="flex items-center justify-center gap-2">
								<svg class="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
									<circle
										class="opacity-25"
										cx="12"
										cy="12"
										r="10"
										stroke="currentColor"
										stroke-width="4"
									></circle>
									<path
										class="opacity-75"
										fill="currentColor"
										d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"
									></path>
								</svg>
								Memuat data...
							</div>
						</td>
					</tr>
				{:else if rows.length === 0}
					<tr>
						<td colspan={columns.length} class="px-4 py-8 text-center text-sm text-slate-500">
							{emptyMessage}
						</td>
					</tr>
				{:else}
					{#each rows as r, i}
						{#if row}
							{@render row(r, i)}
						{:else}
							<tr class="hover:bg-slate-50">
								{#each columns as col}
									<td class="px-4 py-3 text-sm text-slate-700 text-{col.align ?? 'left'}">
										{col.field ? String((r as Record<string, unknown>)[col.field as string] ?? '-') : '-'}
									</td>
								{/each}
							</tr>
						{/if}
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>
