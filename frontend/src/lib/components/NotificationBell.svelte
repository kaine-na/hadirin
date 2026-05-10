<script lang="ts">
	import { Bell } from 'lucide-svelte';
	import { notifications, markAsRead, markAllAsRead } from '$lib/stores/notifications.svelte';
	import NotificationDropdown from './NotificationDropdown.svelte';

	let dropdownOpen = $state(false);

	// Tutup dropdown saat klik di luar
	function onDocClick(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (!target.closest('[data-notification-bell]')) {
			dropdownOpen = false;
		}
	}

	function toggleDropdown() {
		dropdownOpen = !dropdownOpen;
	}
</script>

<svelte:window onclick={onDocClick} />

<div class="relative" data-notification-bell>
	<!-- Bell button dengan badge -->
	<button
		type="button"
		class="relative flex h-9 w-9 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700"
		aria-label="Notifikasi"
		onclick={toggleDropdown}
	>
		<Bell size={18} strokeWidth={1.75} />

		{#if notifications.unreadCount > 0}
			<!-- Badge merah dengan animasi pulse -->
			<span
				class="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold leading-none text-white"
			>
				<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-400 opacity-75"></span>
				<span class="relative">
					{notifications.unreadCount > 99 ? '99+' : notifications.unreadCount}
				</span>
			</span>
		{/if}
	</button>

	<!-- Dropdown -->
	{#if dropdownOpen}
		<NotificationDropdown
			onClose={() => (dropdownOpen = false)}
			onMarkAsRead={markAsRead}
			onMarkAllAsRead={markAllAsRead}
		/>
	{/if}
</div>
