<script lang="ts">
	import {
		Bell,
		Clock,
		CheckCircle,
		XCircle,
		FileText,
		AlertCircle,
		Check
	} from 'lucide-svelte';
	import { notifications } from '$lib/stores/notifications.svelte';
	import type { Notification } from '$lib/api/notifications';

	interface Props {
		onClose: () => void;
		onMarkAsRead: (id: string) => Promise<void>;
		onMarkAllAsRead: () => Promise<void>;
	}

	let { onClose, onMarkAsRead, onMarkAllAsRead }: Props = $props();

	// Icon per tipe notifikasi
	function getIcon(type: string) {
		switch (type) {
			case 'clock_in_reminder':
				return Clock;
			case 'clock_in_confirmed':
				return CheckCircle;
			case 'leave_approved':
				return CheckCircle;
			case 'leave_rejected':
				return XCircle;
			case 'doc_approved':
				return FileText;
			case 'doc_rejected':
				return XCircle;
			case 'leave_reminder':
			case 'doc_reminder':
				return AlertCircle;
			default:
				return Bell;
		}
	}

	// Warna icon per tipe
	function getIconColor(type: string): string {
		switch (type) {
			case 'clock_in_confirmed':
			case 'leave_approved':
			case 'doc_approved':
				return 'text-green-500';
			case 'leave_rejected':
			case 'doc_rejected':
				return 'text-red-500';
			case 'clock_in_reminder':
			case 'leave_reminder':
			case 'doc_reminder':
				return 'text-amber-500';
			default:
				return 'text-blue-500';
		}
	}

	// Link terkait per tipe notifikasi
	function getLink(n: Notification): string | null {
		const meta = n.metadata as Record<string, string> | undefined;
		if (!meta) return null;

		if (meta.leave_id) return `/leaves`;
		if (meta.doc_id) return `/documents`;
		if (meta.action === 'clock_in') return `/attendance`;
		return null;
	}

	// Format waktu relatif
	function relativeTime(dateStr: string): string {
		const diff = Date.now() - new Date(dateStr).getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(diff / 3600000);
		const days = Math.floor(diff / 86400000);

		if (minutes < 1) return 'Baru saja';
		if (minutes < 60) return `${minutes} menit lalu`;
		if (hours < 24) return `${hours} jam lalu`;
		return `${days} hari lalu`;
	}

	async function handleItemClick(n: Notification) {
		if (!n.is_read) {
			await onMarkAsRead(n.id);
		}
		const link = getLink(n);
		if (link) {
			onClose();
			window.location.href = link;
		}
	}
</script>

<div
	class="absolute right-0 top-11 z-50 w-80 overflow-hidden rounded-xl border border-slate-200 bg-white shadow-xl"
>
	<!-- Header -->
	<div class="flex items-center justify-between border-b border-slate-100 px-4 py-3">
		<div class="flex items-center gap-2">
			<Bell size={16} strokeWidth={1.75} class="text-slate-600" />
			<span class="text-sm font-semibold text-slate-800">Notifikasi</span>
			{#if notifications.unreadCount > 0}
				<span class="rounded-full bg-red-100 px-1.5 py-0.5 text-[10px] font-bold text-red-600">
					{notifications.unreadCount}
				</span>
			{/if}
		</div>

		{#if notifications.unreadCount > 0}
			<button
				type="button"
				class="flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-blue-600 transition-colors hover:bg-blue-50"
				onclick={onMarkAllAsRead}
			>
				<Check size={12} strokeWidth={2} />
				Tandai semua dibaca
			</button>
		{/if}
	</div>

	<!-- List notifikasi -->
	<div class="max-h-96 overflow-y-auto">
		{#if notifications.list.length === 0}
			<div class="flex flex-col items-center justify-center gap-2 py-10 text-slate-400">
				<Bell size={32} strokeWidth={1.25} />
				<p class="text-sm">Belum ada notifikasi</p>
			</div>
		{:else}
			{#each notifications.list.slice(0, 10) as n (n.id)}
				{@const Icon = getIcon(n.type)}
				{@const iconColor = getIconColor(n.type)}
				{@const link = getLink(n)}
				<button
					type="button"
					class="flex w-full items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-slate-50
						{n.is_read ? 'bg-white' : 'bg-blue-50'}"
					onclick={() => handleItemClick(n)}
				>
					<!-- Icon tipe -->
					<div class="mt-0.5 shrink-0">
						<Icon size={18} strokeWidth={1.75} class={iconColor} />
					</div>

					<!-- Konten -->
					<div class="min-w-0 flex-1">
						<p class="truncate text-sm font-semibold text-slate-800">{n.title}</p>
						<p class="mt-0.5 line-clamp-2 text-xs text-slate-500">{n.message}</p>
						<div class="mt-1 flex items-center gap-2">
							<span class="text-[10px] text-slate-400">{relativeTime(n.created_at)}</span>
							{#if link}
								<span class="text-[10px] font-medium text-blue-500">Lihat detail →</span>
							{/if}
						</div>
					</div>

					<!-- Dot unread -->
					{#if !n.is_read}
						<div class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-blue-500"></div>
					{/if}
				</button>

				<!-- Divider -->
				<div class="mx-4 border-b border-slate-100 last:hidden"></div>
			{/each}
		{/if}
	</div>

	<!-- Footer -->
	{#if notifications.list.length > 0}
		<div class="border-t border-slate-100 px-4 py-2.5">
			<p class="text-center text-xs text-slate-400">
				Menampilkan {Math.min(10, notifications.list.length)} dari {notifications.list.length} notifikasi
			</p>
		</div>
	{/if}
</div>
