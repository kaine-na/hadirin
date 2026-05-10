import { browser } from '$app/environment';
import { auth } from '$lib/stores/auth.svelte';
import type { Notification } from '$lib/api/notifications';

// --- State reaktif dengan Svelte 5 runes ---

let _notifications = $state<Notification[]>([]);
let _unreadCount = $state(0);
let _connected = $state(false);

/** Daftar notifikasi terbaru (max 50 di memori). */
export const notifications = {
	get list() {
		return _notifications;
	},
	get unreadCount() {
		return _unreadCount;
	},
	get connected() {
		return _connected;
	}
};

// --- SSE connection management ---

let eventSource: EventSource | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let reconnectDelay = 2000; // mulai 2 detik, exponential backoff

/** Inisialisasi SSE dan load notifikasi awal. */
export async function initNotifications() {
	if (!browser) return;
	if (!auth.isLoggedIn) return;

	await loadInitialNotifications();
	connectSSE();
}

/** Bersihkan koneksi SSE saat logout. */
export function destroyNotifications() {
	disconnectSSE();
	_notifications = [];
	_unreadCount = 0;
	_connected = false;
}

/** Tambahkan notifikasi baru ke state (dipanggil dari SSE event). */
function addNotification(n: Notification) {
	// Hindari duplikat
	if (_notifications.some((existing) => existing.id === n.id)) return;

	_notifications = [n, ..._notifications].slice(0, 50);
	if (!n.is_read) {
		_unreadCount += 1;
	}
}

/** Load notifikasi awal dari REST API. */
async function loadInitialNotifications() {
	try {
		const { notificationsApi } = await import('$lib/api/notifications');
		const [listRes, countRes] = await Promise.all([
			notificationsApi.list({ page: 1, page_size: 20 }),
			notificationsApi.getUnreadCount()
		]);
		_notifications = listRes.notifications ?? [];
		_unreadCount = countRes.count ?? 0;
	} catch {
		// Gagal load awal — tidak fatal, SSE akan update nanti
	}
}

/** Buka koneksi SSE ke backend. */
function connectSSE() {
	if (!browser || !auth.token) return;

	disconnectSSE();

	// SSE tidak mendukung custom header, jadi token dikirim via query param
	// Backend harus mendukung ?token= sebagai fallback auth
	// Alternatif: gunakan cookie-based auth untuk SSE
	const url = `/api/notifications/stream`;

	// Gunakan fetch-based SSE dengan Authorization header
	// karena EventSource tidak mendukung custom headers
	startFetchSSE(url);
}

/** SSE menggunakan fetch + ReadableStream agar bisa kirim Authorization header. */
function startFetchSSE(url: string) {
	if (!browser) return;

	const token = auth.token;
	if (!token) return;

	let abortController = new AbortController();

	async function connect() {
		try {
			const response = await fetch(url, {
				headers: {
					Authorization: `Bearer ${token}`,
					Accept: 'text/event-stream'
				},
				signal: abortController.signal
			});

			if (!response.ok || !response.body) {
				scheduleReconnect();
				return;
			}

			_connected = true;
			reconnectDelay = 2000; // reset delay setelah berhasil connect

			const reader = response.body.getReader();
			const decoder = new TextDecoder();
			let buffer = '';

			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				buffer += decoder.decode(value, { stream: true });
				const lines = buffer.split('\n');
				buffer = lines.pop() ?? '';

				let eventType = '';
				let dataLine = '';

				for (const line of lines) {
					if (line.startsWith('event: ')) {
						eventType = line.slice(7).trim();
					} else if (line.startsWith('data: ')) {
						dataLine = line.slice(6).trim();
					} else if (line === '' && dataLine) {
						// Event lengkap
						if (eventType === 'notification') {
							try {
								const n = JSON.parse(dataLine) as Notification;
								addNotification(n);
							} catch {
								// Abaikan parse error
							}
						}
						eventType = '';
						dataLine = '';
					}
				}
			}
		} catch (err: unknown) {
			if (err instanceof Error && err.name === 'AbortError') return;
			// Koneksi putus, coba reconnect
		}

		_connected = false;
		scheduleReconnect();
	}

	// Simpan abort controller untuk cleanup
	(window as unknown as Record<string, unknown>).__sseAbort = abortController;
	connect();
}

function disconnectSSE() {
	if (eventSource) {
		eventSource.close();
		eventSource = null;
	}
	// Abort fetch-based SSE
	const abort = (window as unknown as Record<string, unknown>).__sseAbort as AbortController | undefined;
	if (abort) {
		abort.abort();
		delete (window as unknown as Record<string, unknown>).__sseAbort;
	}
	if (reconnectTimer) {
		clearTimeout(reconnectTimer);
		reconnectTimer = null;
	}
	_connected = false;
}

function scheduleReconnect() {
	if (!browser || !auth.isLoggedIn) return;

	reconnectTimer = setTimeout(() => {
		reconnectDelay = Math.min(reconnectDelay * 2, 30000); // max 30 detik
		connectSSE();
	}, reconnectDelay);
}

// --- Actions ---

/** Tandai satu notifikasi sebagai sudah dibaca. */
export async function markAsRead(id: string) {
	try {
		const { notificationsApi } = await import('$lib/api/notifications');
		await notificationsApi.markAsRead(id);

		_notifications = _notifications.map((n) =>
			n.id === id ? { ...n, is_read: true } : n
		);
		_unreadCount = Math.max(0, _unreadCount - 1);
	} catch {
		// Abaikan error — UI tetap responsif
	}
}

/** Tandai semua notifikasi sebagai sudah dibaca. */
export async function markAllAsRead() {
	try {
		const { notificationsApi } = await import('$lib/api/notifications');
		await notificationsApi.markAllAsRead();

		_notifications = _notifications.map((n) => ({ ...n, is_read: true }));
		_unreadCount = 0;
	} catch {
		// Abaikan error
	}
}
