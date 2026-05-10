import { api } from './client';
import type { Paginated } from '$lib/types';

export type LeaveStatus = 'pending' | 'approved' | 'rejected' | 'cancelled';

export interface LeaveType {
	id: string;
	name: string;
	max_days: number;
	is_paid: boolean;
	description?: string;
	created_at: string;
	updated_at: string;
}

export interface LeaveBalance {
	id: string;
	user_id: string;
	leave_type_id: string;
	leave_type_name: string;
	year: number;
	total_days: number;
	used_days: number;
	remaining_days: number;
	created_at: string;
	updated_at: string;
}

export interface LeaveRequest {
	id: string;
	user_id: string;
	user_name?: string;
	leave_type_id: string;
	leave_type_name?: string;
	start_date: string;
	end_date: string;
	total_days: number;
	reason: string;
	status: LeaveStatus;
	approved_by?: string;
	approved_by_name?: string;
	approved_at?: string;
	rejection_reason?: string;
	created_at: string;
	updated_at: string;
}

export interface CreateLeavePayload {
	leave_type_id: string;
	start_date: string;
	end_date: string;
	reason: string;
}

export interface RejectLeavePayload {
	rejection_reason: string;
}

export interface AIRecommendation {
	recommendation: string;
	reason: string;
}

export interface LeaveQuery {
	user_id?: string;
	status?: string;
	start_date?: string;
	end_date?: string;
	page?: number;
	page_size?: number;
	[key: string]: string | number | boolean | undefined | null;
}

export const leavesApi = {
	// Jenis cuti
	getTypes: () => api.get<LeaveType[]>('/api/leaves/types'),

	// Pengajuan cuti
	create: (payload: CreateLeavePayload) => api.post<LeaveRequest>('/api/leaves', payload),
	list: (q: LeaveQuery = {}) => api.get<Paginated<LeaveRequest>>('/api/leaves', q),
	getById: (id: string) => api.get<LeaveRequest>(`/api/leaves/${id}`),
	approve: (id: string) => api.put<LeaveRequest>(`/api/leaves/${id}/approve`, {}),
	reject: (id: string, payload: RejectLeavePayload) =>
		api.put<LeaveRequest>(`/api/leaves/${id}/reject`, payload),
	cancel: (id: string) => api.put<LeaveRequest>(`/api/leaves/${id}/cancel`, {}),

	// Saldo cuti
	getMyBalance: (year?: number) =>
		api.get<LeaveBalance[]>('/api/leaves/balance', year ? { year } : {}),
	getBalanceByUser: (userId: string, year?: number) =>
		api.get<LeaveBalance[]>(`/api/leaves/balance/${userId}`, year ? { year } : {}),

	// AI recommendation
	getAIRecommendation: (id: string) =>
		api.get<AIRecommendation>(`/api/leaves/${id}/ai-recommendation`)
};
