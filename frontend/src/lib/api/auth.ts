import { api } from './client';
import type { LoginPayload, LoginResponse, User } from '$lib/types';

export const authApi = {
	login: (payload: LoginPayload) => api.post<LoginResponse>('/api/auth/login', payload),
	logout: () => api.post<null>('/api/auth/logout'),
	me: () => api.get<User>('/api/auth/me')
};
