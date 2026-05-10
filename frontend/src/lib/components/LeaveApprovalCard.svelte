<script lang="ts">
	import { leavesApi, type LeaveRequest, type AIRecommendation } from '$lib/api/leaves';
	import LeaveStatusBadge from './LeaveStatusBadge.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { Sparkles, CheckCircle, XCircle, Calendar, User, Clock } from 'lucide-svelte';

	interface Props {
		request: LeaveRequest;
		onUpdate?: () => void;
	}

	let { request, onUpdate }: Props = $props();

	let showRejectModal = $state(false);
	let rejectionReason = $state('');
	let processing = $state(false);
	let aiRec = $state<AIRecommendation | null>(null);
	let loadingAI = $state(false);

	async function loadAIRecommendation() {
		if (aiRec || loadingAI) return;
		loadingAI = true;
		try {
			aiRec = await leavesApi.getAIRecommendation(request.id);
		} catch {
			// AI tidak wajib berhasil
		} finally {
			loadingAI = false;
		}
	}

	async function handleApprove() {
		processing = true;
		try {
			await leavesApi.approve(request.id);
			toast.success('Pengajuan cuti berhasil disetujui');
			onUpdate?.();
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : 'Gagal menyetujui cuti';
			toast.error(msg);
		} finally {
			processing = false;
		}
	}

	async function handleReject() {
		if (!rejectionReason.trim()) {
			toast.error('Alasan penolakan wajib diisi');
			return;
		}
		processing = true;
		try {
			await leavesApi.reject(request.id, { rejection_reason: rejectionReason });
			toast.success('Pengajuan cuti berhasil ditolak');
			showRejectModal = false;
			rejectionReason = '';
			onUpdate?.();
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : 'Gagal menolak cuti';
			toast.error(msg);
		} finally {
			processing = false;
		}
	}

	// Load AI recommendation saat komponen mount (hanya untuk pending)
	$effect(() => {
		if (request.status === 'pending') {
			loadAIRecommendation();
		}
	});
</script>

<div
	class="rounded-xl border border-slate-100 bg-white p-5 shadow-sm transition-all duration-200 hover:shadow-md"
>
	<!-- Header -->
	<div class="flex items-start justify-between gap-3">
		<div class="flex items-center gap-3">
			<div
				class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-blue-100 text-sm font-semibold text-blue-700"
			>
				{request.user_name?.charAt(0)?.toUpperCase() ?? '?'}
			</div>
			<div>
				<p class="font-semibold text-slate-900">{request.user_name ?? 'Karyawan'}</p>
				<p class="text-xs text-slate-500">{request.leave_type_name ?? 'Cuti'}</p>
			</div>
		</div>
		<LeaveStatusBadge status={request.status} />
	</div>

	<!-- Detail -->
	<div class="mt-4 grid grid-cols-2 gap-3 text-sm">
		<div class="flex items-center gap-2 text-slate-600">
			<Calendar size={14} strokeWidth={1.75} class="shrink-0 text-slate-400" />
			<span>
				{new Date(request.start_date).toLocaleDateString('id-ID', {
					day: 'numeric',
					month: 'short',
					year: 'numeric'
				})}
				{#if request.start_date !== request.end_date}
					→ {new Date(request.end_date).toLocaleDateString('id-ID', {
						day: 'numeric',
						month: 'short',
						year: 'numeric'
					})}
				{/if}
			</span>
		</div>
		<div class="flex items-center gap-2 text-slate-600">
			<Clock size={14} strokeWidth={1.75} class="shrink-0 text-slate-400" />
			<span>{request.total_days} hari</span>
		</div>
	</div>

	<!-- Alasan -->
	<div class="mt-3 rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-700">
		<span class="font-medium text-slate-500">Alasan: </span>{request.reason}
	</div>

	<!-- Rejection reason (jika ditolak) -->
	{#if request.status === 'rejected' && request.rejection_reason}
		<div class="mt-2 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
			<span class="font-medium">Alasan penolakan: </span>{request.rejection_reason}
		</div>
	{/if}

	<!-- AI Recommendation Badge -->
	{#if request.status === 'pending'}
		<div class="mt-3">
			{#if loadingAI}
				<div class="flex items-center gap-2 text-xs text-slate-400">
					<span class="h-3 w-3 animate-spin rounded-full border border-slate-300 border-t-slate-600"
					></span>
					Memuat rekomendasi AI...
				</div>
			{:else if aiRec}
				<div
					class="flex items-start gap-2 rounded-lg px-3 py-2 text-xs
					{aiRec.recommendation === 'Direkomendasikan disetujui'
						? 'bg-green-50 text-green-700'
						: 'bg-amber-50 text-amber-700'}"
				>
					<Sparkles size={13} strokeWidth={1.75} class="mt-0.5 shrink-0" />
					<div>
						<span class="font-semibold">{aiRec.recommendation}</span>
						{#if aiRec.reason}
							<span class="ml-1 opacity-80">— {aiRec.reason}</span>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Approved by info -->
	{#if request.approved_by_name && request.status !== 'pending'}
		<div class="mt-2 flex items-center gap-1 text-xs text-slate-500">
			<User size={12} strokeWidth={1.75} />
			<span>
				{request.status === 'approved' ? 'Disetujui' : 'Ditolak'} oleh {request.approved_by_name}
			</span>
		</div>
	{/if}

	<!-- Action buttons (hanya untuk pending) -->
	{#if request.status === 'pending'}
		<div class="mt-4 flex gap-2">
			<button
				onclick={handleApprove}
				disabled={processing}
				class="flex flex-1 items-center justify-center gap-2 rounded-lg bg-green-600 px-3 py-2 text-sm font-medium text-white transition-all hover:bg-green-700 active:scale-95 disabled:opacity-50"
			>
				<CheckCircle size={15} strokeWidth={2} />
				Setujui
			</button>
			<button
				onclick={() => (showRejectModal = true)}
				disabled={processing}
				class="flex flex-1 items-center justify-center gap-2 rounded-lg bg-red-600 px-3 py-2 text-sm font-medium text-white transition-all hover:bg-red-700 active:scale-95 disabled:opacity-50"
			>
				<XCircle size={15} strokeWidth={2} />
				Tolak
			</button>
		</div>
	{/if}
</div>

<!-- Reject Modal -->
{#if showRejectModal}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
		role="dialog"
		aria-modal="true"
		aria-labelledby="reject-modal-title"
	>
		<div class="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
			<h3 id="reject-modal-title" class="text-base font-semibold text-slate-900">Tolak Pengajuan Cuti</h3>
			<p class="mt-1 text-sm text-slate-500">
				Berikan alasan penolakan untuk {request.user_name}.
			</p>
			<textarea
				bind:value={rejectionReason}
				rows={3}
				placeholder="Alasan penolakan..."
				class="mt-3 block w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-sm focus:border-red-500 focus:outline-none focus:ring-1 focus:ring-red-500"
			></textarea>
			<div class="mt-4 flex justify-end gap-3">
				<button
					onclick={() => {
						showRejectModal = false;
						rejectionReason = '';
					}}
					disabled={processing}
					class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50"
				>
					Batal
				</button>
				<button
					onclick={handleReject}
					disabled={processing || !rejectionReason.trim()}
					class="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
				>
					{#if processing}Menolak...{:else}Konfirmasi Tolak{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
