<script lang="ts">
	import { AlertTriangle, CheckCircle, Clock, MapPin, Smartphone, Eye, ChevronDown, ChevronUp } from 'lucide-svelte';
	import type { FraudLog } from '$lib/api/fraud';
	import { fraudApi } from '$lib/api/fraud';
	import { toast } from '$lib/stores/toast.svelte';

	interface Props {
		log: FraudLog;
		onUpdate?: () => void;
	}

	let { log, onUpdate }: Props = $props();

	let expanded = $state(false);
	let reviewing = $state(false);
	let reviewNotes = $state('');

	const severityConfig: Record<string, { color: string; bg: string; label: string }> = {
		low: { color: 'text-blue-700', bg: 'bg-blue-100', label: 'Rendah' },
		medium: { color: 'text-amber-700', bg: 'bg-amber-100', label: 'Sedang' },
		high: { color: 'text-orange-700', bg: 'bg-orange-100', label: 'Tinggi' },
		critical: { color: 'text-red-700', bg: 'bg-red-100', label: 'Kritis' }
	};

	const fraudTypeLabel: Record<string, string> = {
		gps_accuracy: 'Akurasi GPS Buruk',
		mock_location: 'Lokasi Palsu',
		velocity_check: 'Velocity Anomaly',
		anomaly_time: 'Waktu Tidak Normal',
		anomaly_location: 'Lokasi Tidak Normal',
		anomaly_device: 'Device Berbeda',
		liveness_fail: 'Selfie Tidak Valid'
	};

	const fraudTypeIcon: Record<string, typeof AlertTriangle> = {
		gps_accuracy: MapPin,
		mock_location: MapPin,
		velocity_check: MapPin,
		anomaly_time: Clock,
		anomaly_location: MapPin,
		anomaly_device: Smartphone,
		liveness_fail: Eye
	};

	const statusConfig: Record<string, { color: string; bg: string; label: string }> = {
		pending: { color: 'text-amber-700', bg: 'bg-amber-100', label: 'Menunggu Review' },
		dismissed: { color: 'text-slate-600', bg: 'bg-slate-100', label: 'Diabaikan' },
		confirmed: { color: 'text-red-700', bg: 'bg-red-100', label: 'Dikonfirmasi' }
	};

	const sev = $derived(severityConfig[log.severity] ?? severityConfig.medium);
	const statusCfg = $derived(statusConfig[log.status] ?? statusConfig.pending);
	const FraudIcon = $derived(fraudTypeIcon[log.fraud_type] ?? AlertTriangle);

	async function dismiss() {
		reviewing = true;
		try {
			await fraudApi.dismissLog(log.id, reviewNotes);
			toast.success('Fraud log berhasil di-dismiss');
			onUpdate?.();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal dismiss');
		} finally {
			reviewing = false;
		}
	}

	async function confirm() {
		reviewing = true;
		try {
			await fraudApi.confirmLog(log.id, reviewNotes);
			toast.success('Fraud log berhasil dikonfirmasi');
			onUpdate?.();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal konfirmasi');
		} finally {
			reviewing = false;
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleString('id-ID', {
			day: 'numeric',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm transition-shadow hover:shadow-md">
	<!-- Header card -->
	<div class="flex items-start gap-4 p-4">
		<!-- Foto thumbnail -->
		{#if log.photo_url}
			<div class="h-14 w-14 shrink-0 overflow-hidden rounded-lg border border-slate-200 bg-slate-100">
				<img
					src={log.photo_url}
					alt="Foto selfie {log.employee_name}"
					class="h-full w-full object-cover"
					loading="lazy"
				/>
			</div>
		{:else}
			<div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-lg border border-slate-200 bg-slate-100">
				<FraudIcon size={24} strokeWidth={1.5} class="text-slate-400" />
			</div>
		{/if}

		<!-- Info utama -->
		<div class="min-w-0 flex-1">
			<div class="flex flex-wrap items-start justify-between gap-2">
				<div>
					<p class="font-semibold text-slate-900">{log.employee_name || 'Karyawan'}</p>
					<p class="text-sm text-slate-500">{fraudTypeLabel[log.fraud_type] ?? log.fraud_type}</p>
				</div>
				<div class="flex flex-wrap items-center gap-2">
					<!-- Severity badge -->
					<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {sev.bg} {sev.color}">
						{sev.label}
					</span>
					<!-- Status badge -->
					<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {statusCfg.bg} {statusCfg.color}">
						{statusCfg.label}
					</span>
				</div>
			</div>

			<p class="mt-1.5 text-sm text-slate-600">{log.description}</p>

			<div class="mt-2 flex items-center gap-1 text-xs text-slate-400">
				<Clock size={12} strokeWidth={2} />
				<span>{formatDate(log.created_at)}</span>
			</div>
		</div>

		<!-- Toggle expand -->
		<button
			type="button"
			onclick={() => (expanded = !expanded)}
			class="shrink-0 rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
			aria-label={expanded ? 'Sembunyikan detail' : 'Lihat detail'}
		>
			{#if expanded}
				<ChevronUp size={16} strokeWidth={2} />
			{:else}
				<ChevronDown size={16} strokeWidth={2} />
			{/if}
		</button>
	</div>

	<!-- Detail yang bisa di-expand -->
	{#if expanded}
		<div class="border-t border-slate-100 bg-slate-50 px-4 py-4 space-y-4">
			<!-- AI Analysis -->
			{#if log.ai_analysis}
				<div>
					<p class="mb-1.5 text-xs font-semibold uppercase tracking-wide text-slate-500">Analisis AI</p>
					<p class="text-sm text-slate-700 leading-relaxed">{log.ai_analysis}</p>
				</div>
			{/if}

			<!-- Evidence -->
			{#if log.evidence && Object.keys(log.evidence).length > 0}
				<div>
					<p class="mb-1.5 text-xs font-semibold uppercase tracking-wide text-slate-500">Bukti</p>
					<div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
						{#each Object.entries(log.evidence) as [key, value]}
							<div class="rounded-lg bg-white border border-slate-200 px-3 py-2">
								<p class="text-xs text-slate-400">{key.replace(/_/g, ' ')}</p>
								<p class="text-sm font-medium text-slate-800 truncate">{String(value)}</p>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Review notes input (hanya jika pending) -->
			{#if log.status === 'pending'}
				<div>
					<label for="review-notes-{log.id}" class="mb-1.5 block text-xs font-semibold uppercase tracking-wide text-slate-500">
						Catatan Review (opsional)
					</label>
					<textarea
						id="review-notes-{log.id}"
						bind:value={reviewNotes}
						rows={2}
						placeholder="Tambahkan catatan..."
						class="w-full resize-none rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-800 placeholder-slate-400 focus:border-blue-400 focus:outline-none focus:ring-2 focus:ring-blue-100"
					></textarea>
					<div class="mt-3 flex gap-3">
						<button
							type="button"
							onclick={dismiss}
							disabled={reviewing}
							class="flex items-center gap-2 rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:opacity-50"
						>
							{#if reviewing}
								<span class="h-3.5 w-3.5 animate-spin rounded-full border-2 border-slate-400 border-t-transparent"></span>
							{/if}
							Abaikan (False Positive)
						</button>
						<button
							type="button"
							onclick={confirm}
							disabled={reviewing}
							class="flex items-center gap-2 rounded-lg bg-red-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-red-700 disabled:opacity-50"
						>
							{#if reviewing}
								<span class="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
							{/if}
							Konfirmasi Fraud
						</button>
					</div>
				</div>
			{:else if log.review_notes}
				<div>
					<p class="mb-1 text-xs font-semibold uppercase tracking-wide text-slate-500">Catatan Review</p>
					<p class="text-sm text-slate-700">{log.review_notes}</p>
				</div>
			{/if}
		</div>
	{/if}
</div>
