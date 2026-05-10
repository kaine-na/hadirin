<script lang="ts">
	import { Sparkles, Loader2, RefreshCw } from 'lucide-svelte';
	import { analyticsApi } from '$lib/api/analytics';
	import { toast } from '$lib/stores/toast.svelte';
	import type { AnalyticsFilter } from '$lib/api/analytics';

	interface Props {
		filter: AnalyticsFilter;
	}

	let { filter }: Props = $props();

	let loading = $state(false);
	let summary = $state('');
	let generatedAt = $state('');
	let displayedText = $state('');
	let typingInterval: ReturnType<typeof setInterval> | null = null;

	function typeText(text: string) {
		displayedText = '';
		let i = 0;
		if (typingInterval) clearInterval(typingInterval);

		typingInterval = setInterval(() => {
			if (i < text.length) {
				displayedText += text[i];
				i++;
			} else {
				if (typingInterval) clearInterval(typingInterval);
			}
		}, 12);
	}

	async function generate() {
		loading = true;
		try {
			const result = await analyticsApi.getExecutiveSummary(filter);
			summary = result.summary;
			generatedAt = new Date(result.generated_at).toLocaleString('id-ID', {
				day: '2-digit',
				month: 'long',
				year: 'numeric',
				hour: '2-digit',
				minute: '2-digit'
			});
			typeText(summary);
		} catch (err) {
			const msg = err instanceof Error ? err.message : 'Gagal generate summary';
			toast.error(msg);
		} finally {
			loading = false;
		}
	}

	import { onDestroy } from 'svelte';
	onDestroy(() => {
		if (typingInterval) clearInterval(typingInterval);
	});
</script>

<div class="rounded-xl bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 p-6 shadow-md ring-1 ring-blue-100">
	<div class="mb-4 flex items-center justify-between">
		<div class="flex items-center gap-2">
			<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-600 text-white">
				<Sparkles size={16} />
			</div>
			<div>
				<h3 class="text-sm font-semibold text-slate-900">AI Executive Summary</h3>
				<p class="text-xs text-slate-500">Ringkasan eksekutif berbasis AI</p>
			</div>
		</div>
		<button
			type="button"
			onclick={generate}
			disabled={loading}
			class="flex items-center gap-1.5 rounded-lg border border-blue-200 bg-white px-3 py-1.5 text-xs font-medium text-blue-700 transition-colors hover:bg-blue-50 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-offset-1 disabled:cursor-not-allowed disabled:opacity-60"
		>
			{#if loading}
				<Loader2 size={13} class="animate-spin" />
				<span>Generating...</span>
			{:else}
				<RefreshCw size={13} />
				<span>{summary ? 'Regenerate' : 'Generate'}</span>
			{/if}
		</button>
	</div>

	{#if loading && !displayedText}
		<div class="flex items-center gap-2 py-6 text-sm text-slate-500">
			<Loader2 size={16} class="animate-spin text-blue-500" />
			<span>AI sedang menganalisis data kehadiran...</span>
		</div>
	{:else if displayedText}
		<div class="space-y-2">
			<p class="whitespace-pre-wrap text-sm italic leading-relaxed text-slate-700">
				{displayedText}
				{#if loading}<span class="animate-pulse">▌</span>{/if}
			</p>
			{#if generatedAt && !loading}
				<p class="text-right text-xs text-slate-400">Dibuat pada {generatedAt}</p>
			{/if}
		</div>
	{:else}
		<div class="flex flex-col items-center gap-2 py-6 text-center">
			<Sparkles size={24} class="text-blue-300" />
			<p class="text-sm text-slate-500">Klik "Generate" untuk membuat ringkasan eksekutif AI</p>
			<p class="text-xs text-slate-400">AI akan menganalisis data kehadiran dan memberikan insight</p>
		</div>
	{/if}
</div>
