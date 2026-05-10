export type ToastKind = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
	id: number;
	kind: ToastKind;
	message: string;
}

let _toasts = $state<Toast[]>([]);
let _nextId = 0;

export const toasts = {
	get list() {
		return _toasts;
	}
};

export function pushToast(kind: ToastKind, message: string, durationMs = 4000) {
	const id = ++_nextId;
	_toasts = [..._toasts, { id, kind, message }];
	if (durationMs > 0) {
		setTimeout(() => dismissToast(id), durationMs);
	}
}

export function dismissToast(id: number) {
	_toasts = _toasts.filter((t) => t.id !== id);
}

export const toast = {
	success: (msg: string) => pushToast('success', msg),
	error: (msg: string) => pushToast('error', msg),
	info: (msg: string) => pushToast('info', msg),
	warning: (msg: string) => pushToast('warning', msg)
};
