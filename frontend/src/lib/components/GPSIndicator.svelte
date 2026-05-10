<script lang="ts">
	import { MapPin, AlertTriangle, CheckCircle, WifiOff } from 'lucide-svelte';

	interface Props {
		accuracy: number | null;
		latitude?: number | null;
		longitude?: number | null;
		loading?: boolean;
	}

	let { accuracy, latitude = null, longitude = null, loading = false }: Props = $props();

	// Tentukan status berdasarkan akurasi GPS
	const status = $derived(() => {
		if (loading) return 'loading';
		if (accuracy === null) return 'unavailable';
		if (accuracy <= 20) return 'excellent';
		if (accuracy <= 50) return 'good';
		if (accuracy <= 100) return 'fair';
		return 'poor';
	});

	const statusConfig = $derived(() => {
		switch (status()) {
			case 'excellent':
				return {
					color: 'text-emerald-600',
					bg: 'bg-emerald-50',
					border: 'border-emerald-200',
					dot: 'bg-emerald-500',
					label: 'Sangat Akurat',
					sublabel: `±${accuracy?.toFixed(0)}m`
				};
			case 'good':
				return {
					color: 'text-blue-600',
					bg: 'bg-blue-50',
					border: 'border-blue-200',
					dot: 'bg-blue-500',
					label: 'Akurat',
					sublabel: `±${accuracy?.toFixed(0)}m`
				};
			case 'fair':
				return {
					color: 'text-amber-600',
					bg: 'bg-amber-50',
					border: 'border-amber-200',
					dot: 'bg-amber-500',
					label: 'Cukup Akurat',
					sublabel: `±${accuracy?.toFixed(0)}m`
				};
			case 'poor':
				return {
					color: 'text-red-600',
					bg: 'bg-red-50',
					border: 'border-red-200',
					dot: 'bg-red-500',
					label: 'Akurasi Buruk',
					sublabel: `±${accuracy?.toFixed(0)}m — GPS mungkin diblokir`
				};
			case 'unavailable':
				return {
					color: 'text-slate-500',
					bg: 'bg-slate-50',
					border: 'border-slate-200',
					dot: 'bg-slate-400',
					label: 'GPS Tidak Tersedia',
					sublabel: 'Aktifkan lokasi di perangkat Anda'
				};
			default:
				return {
					color: 'text-slate-400',
					bg: 'bg-slate-50',
					border: 'border-slate-200',
					dot: 'bg-slate-300',
					label: 'Memuat GPS...',
					sublabel: ''
				};
		}
	});

	const isBlocked = $derived(status() === 'poor');
</script>

<div
	class="flex items-center gap-3 rounded-xl border px-4 py-3 {statusConfig().bg} {statusConfig().border}"
	role="status"
	aria-label="Status GPS: {statusConfig().label}"
>
	<!-- Ikon -->
	<div class="shrink-0 {statusConfig().color}">
		{#if status() === 'loading'}
			<MapPin size={18} strokeWidth={2} class="animate-pulse" />
		{:else if status() === 'unavailable'}
			<WifiOff size={18} strokeWidth={2} />
		{:else if isBlocked}
			<AlertTriangle size={18} strokeWidth={2} />
		{:else}
			<CheckCircle size={18} strokeWidth={2} />
		{/if}
	</div>

	<!-- Teks -->
	<div class="min-w-0 flex-1">
		<div class="flex items-center gap-2">
			<!-- Dot indikator -->
			<span
				class="inline-block h-2 w-2 shrink-0 rounded-full {statusConfig().dot} {status() === 'loading' ? 'animate-pulse' : ''}"
				aria-hidden="true"
			></span>
			<span class="text-sm font-medium {statusConfig().color}">{statusConfig().label}</span>
		</div>
		{#if statusConfig().sublabel}
			<p class="mt-0.5 text-xs {statusConfig().color} opacity-75">{statusConfig().sublabel}</p>
		{/if}
	</div>

	<!-- Koordinat (jika tersedia) -->
	{#if latitude && longitude && status() !== 'poor'}
		<div class="shrink-0 text-right">
			<p class="text-xs text-slate-500">{latitude.toFixed(4)}</p>
			<p class="text-xs text-slate-500">{longitude.toFixed(4)}</p>
		</div>
	{/if}
</div>

<!-- Warning jika GPS buruk -->
{#if isBlocked}
	<div class="mt-2 flex items-start gap-2 rounded-lg bg-red-50 px-3 py-2.5 text-xs text-red-700">
		<AlertTriangle size={14} strokeWidth={2} class="mt-0.5 shrink-0" />
		<span>
			Akurasi GPS terlalu rendah untuk clock-in. Pastikan GPS aktif dan Anda berada di area dengan
			sinyal yang baik.
		</span>
	</div>
{/if}
