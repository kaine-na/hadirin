<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import { auth, setAuth, clearAuth } from '$lib/stores/auth.svelte';
	import { authApi } from '$lib/api';
	import { initNotifications, destroyNotifications } from '$lib/stores/notifications.svelte';

	let { children } = $props();
	let ready = $state(false);

	onMount(async () => {
		if (!auth.isLoggedIn) {
			await goto('/login');
			return;
		}
		// Validasi token dengan hit /me; kalau gagal, redirect.
		try {
			const user = await authApi.me();
			setAuth(user, auth.token!);
			ready = true;
			// Inisialisasi notification store setelah auth valid
			await initNotifications();
		} catch {
			clearAuth();
			await goto('/login');
		}
	});

	onDestroy(() => {
		destroyNotifications();
	});
</script>

{#if ready}
	<div class="flex min-h-screen bg-slate-50">
		<Sidebar />
		<div class="flex min-w-0 flex-1 flex-col">
			<Navbar />
			<main class="flex-1 overflow-x-hidden px-4 py-6 lg:px-8">
				{@render children?.()}
			</main>
		</div>
	</div>
{:else}
	<div class="flex min-h-screen items-center justify-center">
		<div class="flex items-center gap-3 text-slate-500">
			<svg class="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
				></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"
				></path>
			</svg>
			<span class="text-sm">Memuat sesi...</span>
		</div>
	</div>
{/if}
