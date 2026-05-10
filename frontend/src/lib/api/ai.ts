import { api } from './client';
import type { AIReport, AnalyzePayload } from '$lib/types';

export const aiApi = {
	analyze: (employeeId: string, payload: AnalyzePayload) =>
		api.post<AIReport>(`/api/ai/analyze/${employeeId}`, payload),
	listReports: (employeeId: string) => api.get<AIReport[]>(`/api/ai/reports/${employeeId}`),
	getReport: (id: string) => api.get<AIReport>(`/api/ai/report/${id}`)
};
