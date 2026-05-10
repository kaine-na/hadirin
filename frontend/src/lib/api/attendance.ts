import { api } from './client';
import type { Attendance, AttendanceOverridePayload, Paginated } from '$lib/types';

export interface AttendanceQuery {
	start_date?: string;
	end_date?: string;
	page?: number;
	page_size?: number;
	[key: string]: string | number | boolean | undefined | null;
}

export const attendanceApi = {
	clockIn: (notes?: string) => api.post<Attendance>('/api/attendance/clock-in', { notes: notes ?? '' }),
	clockOut: (notes?: string) =>
		api.post<Attendance>('/api/attendance/clock-out', { notes: notes ?? '' }),
	today: () => api.get<Attendance | null>('/api/attendance/today'),
	me: (q: AttendanceQuery = {}) => api.get<Paginated<Attendance>>('/api/attendance/me', q),
	byEmployee: (employeeId: string, q: AttendanceQuery = {}) =>
		api.get<Paginated<Attendance>>(`/api/attendance/${employeeId}`, q),
	override: (id: string, payload: AttendanceOverridePayload) =>
		api.put<Attendance>(`/api/attendance/${id}`, payload),
	exportCsv: (q: AttendanceQuery = {}) => api.download('/api/attendance/export/csv', q)
};
