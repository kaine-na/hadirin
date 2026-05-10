<script lang="ts">
	import { Download, Loader2 } from 'lucide-svelte';
	import { analyticsApi } from '$lib/api/analytics';
	import { toast } from '$lib/stores/toast.svelte';
	import type { AnalyticsFilter } from '$lib/api/analytics';

	interface Props {
		filter: AnalyticsFilter;
	}

	let { filter }: Props = $props();

	let loading = $state(false);

	async function handleExport() {
		loading = true;
		try {
			await analyticsApi.exportPDF(filter);
			toast.success('PDF berhasil diunduh');
		} catch (err) {
			const msg = err instanceof Error ? err.message : 'Gagal mengunduh PDF';
			toast.error(msg);
		} finally {
			loading = false;
		}
	}
</script>

<button
	type="button"
	onclick={handleExport}
	disabled={loading}
	class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 disabled:cursor-not-allowed disabled:opacity-60"
>
	{#if loading}
		<Loader2 size={16} class="animate-spin" />
		<span>Membuat PDF...</span>
	{:else}
		<Download size={16} />
		<span>Export PDF</span>
	{/if}
</button>
