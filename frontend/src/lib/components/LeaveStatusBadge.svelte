<script lang="ts">
	type Status = 'pending' | 'approved' | 'rejected' | 'cancelled';

	interface Props {
		status: Status;
		size?: 'sm' | 'md';
	}

	let { status, size = 'md' }: Props = $props();

	const config: Record<Status, { label: string; classes: string }> = {
		pending: {
			label: 'Menunggu',
			classes: 'bg-amber-100 text-amber-700 border border-amber-200'
		},
		approved: {
			label: 'Disetujui',
			classes: 'bg-green-100 text-green-700 border border-green-200'
		},
		rejected: {
			label: 'Ditolak',
			classes: 'bg-red-100 text-red-700 border border-red-200'
		},
		cancelled: {
			label: 'Dibatalkan',
			classes: 'bg-slate-100 text-slate-600 border border-slate-200'
		}
	};

	const current = $derived(config[status] ?? config.pending);
	const sizeClass = $derived(size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-xs');
</script>

<span class="inline-flex items-center rounded-full font-medium {sizeClass} {current.classes}">
	{current.label}
</span>
