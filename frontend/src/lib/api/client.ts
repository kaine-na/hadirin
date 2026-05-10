import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { auth, clearAuth } from '$lib/stores/auth.svelte';
import type { ApiResponse } from '$lib/types';

/**
 * Base URL untuk API backend. Saat dev, Vite mem-proxy /api ke backend.
 * Di production, set PUBLIC_API_BASE_URL jika backend berbeda origin.
 */
const API_BASE = '';

export class ApiError extends Error {
	status: number;
	body?: unknown;

	constructor(message: string, status: number, body?: unknown) {
		super(message);
		this.name = 'ApiError';
		this.status = status;
		this.body = body;
	}
}

interface RequestOptions {
	method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
	body?: unknown;
	query?: Record<string, string | number | boolean | undefined | null>;
	headers?: Record<string, string>;
	/** Jika true, body dikirim apa adanya (FormData). */
	rawBody?: boolean;
	/** Jika true, return Blob instead of JSON (untuk download). */
	asBlob?: boolean;
}

function buildURL(path: string, query?: RequestOptions['query']): string {
	const url = `${API_BASE}${path}`;
	if (!query) return url;

	const params = new URLSearchParams();
	for (const [key, value] of Object.entries(query)) {
		if (value === undefined || value === null || value === '') continue;
		params.append(key, String(value));
	}
	const qs = params.toString();
	return qs ? `${url}?${qs}` : url;
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
	const headers: Record<string, string> = {
		Accept: 'application/json',
		...(opts.headers ?? {})
	};

	if (auth.token) {
		headers.Authorization = `Bearer ${auth.token}`;
	}

	let body: BodyInit | undefined;
	if (opts.body !== undefined && opts.body !== null) {
		if (opts.rawBody) {
			body = opts.body as BodyInit;
		} else {
			headers['Content-Type'] = 'application/json';
			body = JSON.stringify(opts.body);
		}
	}

	const url = buildURL(path, opts.query);
	const res = await fetch(url, {
		method: opts.method ?? 'GET',
		headers,
		body
	});

	if (res.status === 401) {
		clearAuth();
		if (browser) {
			await goto('/login');
		}
		throw new ApiError('Sesi berakhir, silakan login kembali', 401);
	}

	if (opts.asBlob) {
		if (!res.ok) {
			const text = await res.text().catch(() => '');
			throw new ApiError(text || 'Gagal mengunduh file', res.status);
		}
		return (await res.blob()) as unknown as T;
	}

	const contentType = res.headers.get('Content-Type') ?? '';
	let payload: ApiResponse<T> | null = null;
	if (contentType.includes('application/json')) {
		payload = (await res.json()) as ApiResponse<T>;
	}

	if (!res.ok) {
		const message = payload?.message || `Request gagal (${res.status})`;
		throw new ApiError(message, res.status, payload);
	}

	if (!payload) {
		return null as unknown as T;
	}

	if (!payload.success) {
		throw new ApiError(payload.message || 'Operasi gagal', res.status, payload);
	}

	return (payload.data as T) ?? (null as unknown as T);
}

export const api = {
	get: <T>(path: string, query?: RequestOptions['query']) => request<T>(path, { method: 'GET', query }),
	post: <T>(path: string, body?: unknown) => request<T>(path, { method: 'POST', body }),
	put: <T>(path: string, body?: unknown) => request<T>(path, { method: 'PUT', body }),
	del: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
	upload: <T>(path: string, form: FormData) =>
		request<T>(path, { method: 'POST', body: form, rawBody: true }),
	download: (path: string, query?: RequestOptions['query']) =>
		request<Blob>(path, { method: 'GET', query, asBlob: true })
};

/**
 * Helper untuk download blob sebagai file.
 */
export function downloadBlob(blob: Blob, filename: string) {
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = filename;
	document.body.appendChild(a);
	a.click();
	a.remove();
	URL.revokeObjectURL(url);
}
