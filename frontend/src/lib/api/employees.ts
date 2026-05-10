import { api } from './client';
import type {
	CreateEmployeePayload,
	Employee,
	Paginated,
	UpdateEmployeePayload
} from '$lib/types';

export interface EmployeeListQuery {
	department?: string;
	role?: string;
	search?: string;
	page?: number;
	page_size?: number;
	[key: string]: string | number | boolean | undefined | null;
}

export const employeesApi = {
	list: (q: EmployeeListQuery = {}) => api.get<Paginated<Employee>>('/api/employees', q),
	get: (id: string) => api.get<Employee>(`/api/employees/${id}`),
	create: (payload: CreateEmployeePayload) => api.post<Employee>('/api/employees', payload),
	update: (id: string, payload: UpdateEmployeePayload) =>
		api.put<Employee>(`/api/employees/${id}`, payload),
	remove: (id: string) => api.del<null>(`/api/employees/${id}`),
	uploadPhoto: (id: string, file: File) => {
		const form = new FormData();
		form.append('photo', file);
		return api.upload<{ photo_url: string }>(`/api/employees/${id}/photo`, form);
	}
};
