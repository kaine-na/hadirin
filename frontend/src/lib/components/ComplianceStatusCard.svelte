<script lang="ts">
	import { ShieldCheck, AlertTriangle, CheckCircle2, Clock } from 'lucide-svelte';

	interface Props {
		status: 'green' | 'yellow' | 'red';
		title?: string;
		description?: string;
		pending?: number;
		done?: number;
		overdue?: number;
	}

	let {
		status,
		title = 'Status Kepatuhan',
		description = '',
		pending = 0,
		done = 0,
		overdue = 0
	}: Props = $props();

	const config = $derived(() => {
		switch (status) {
			case 'green':
				return {
					bg: 'bg-emerald-50',
					border: 'border-emerald-200',
					iconBg: 'bg-emerald-100',
					iconColor: 'text-emerald-600',
					titleColor: 'text-emerald-900',
					descColor: 'text-emerald-700',
					label: 'Semua Oke',
					icon: CheckCircle2
				};
			case 'yellow':
				return {
					bg: 'bg-amber-50',
					border: 'border-amber-200',
					iconBg: 'bg-amber-100',
					iconColor: 'text-amber-600',
					titleColor: 'text-amber-900',
					descColor: 'text-amber-700',
					label: 'Perlu Perhatian',
					icon: Clock
				};
			case 'red':
				return {
					bg: 'bg-red-50',
					border: 'border-red-200',
					iconBg: 'bg-red-100',
					iconColor: 'text-red-600',
					titleColor: 'text-red-900',
					descColor: 'text-red-700',
					label: 'Ada yang Terlambat',
					icon: AlertTriangle
				};
		}
	});
</script>

<div class="rounded-xl border p-5 {config().bg} {config().border}">
	<div class="flex items-start gap-4">
		<!-- Icon -->
		<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl {config().iconBg}">
			{#if status === 'green'}
				<ShieldCheck size={20} class={config().iconColor} strokeWidth={1.75} />
			{:else if status === 'yellow'}
				<Clock size={20} class={config().iconColor} strokeWidth={1.75} />
			{:else}
				<AlertTriangle size={20} class={config().iconColor} strokeWidth={1.75} />
			{/if}
		</div>

		<!-- Konten -->
		<div class="flex-1">
			<div class="flex items-center justify-between">
				<p class="text-sm font-semibold {config().titleColor}">{title}</p>
				<span class="rounded-full px-2.5 py-0.5 text-xs font-medium {config().iconBg} {config().iconColor}">
					{config().label}
				</span>
			</div>
			{#if description}
				<p class="mt-0.5 text-xs {config().descColor}">{description}</p>
			{/if}

			<!-- Statistik mini -->
			<div class="mt-3 flex items-center gap-4">
				<div class="flex items-center gap-1.5 text-xs">
					<span class="h-2 w-2 rounded-full bg-emerald-500"></span>
					<span class="{config().descColor}">{done} selesai</span>
				</div>
				<div class="flex items-center gap-1.5 text-xs">
					<span class="h-2 w-2 rounded-full bg-amber-500"></span>
					<span class="{config().descColor}">{pending} pending</span>
				</div>
				{#if overdue > 0}
					<div class="flex items-center gap-1.5 text-xs">
						<span class="h-2 w-2 rounded-full bg-red-500"></span>
						<span class="{config().descColor}">{overdue} terlambat</span>
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>
