import { api } from './client';

export interface Notification {
	id: string;
	user_id: string;
	type: string;
	title: string;
	message: string;
	is_read: boolean;
	metadata?: Record<string, unknown>;
	created_at: string;
}

export interface NotificationListResponse {
	notifications: Notification[];
	total: number;
	page: number;
	page_size: number;
}

export interface UnreadCountResponse {
	count: number;
}

export const notificationsApi = {
	/** Ambil daftar notifikasi (paginated). */
	list(params?: { page?: number; page_size?: number; unread_only?: boolean }) {
		return api.get<NotificationListResponse>('/api/notifications', params);
	},

	/** Ambil jumlah notifikasi yang belum dibaca. */
	getUnreadCount() {
		return api.get<UnreadCountResponse>('/api/notifications/unread-count');
	},

	/** Tandai satu notifikasi sebagai sudah dibaca. */
	markAsRead(id: string) {
		return api.put<void>(`/api/notifications/${id}/read`);
	},

	/** Tandai semua notifikasi sebagai sudah dibaca. */
	markAllAsRead() {
		return api.put<void>('/api/notifications/read-all');
	}
};
