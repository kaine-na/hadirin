import { api, downloadBlob } from './client';

export interface AttendanceSummary {
	total_employees: number;
	total_working_days: number;
	total_present: number;
	total_late: number;
	total_absent: number;
	total_leave: number;
	total_sick: number;
	attendance_rate: number;
	lateness_rate: number;
	period_start: string;
	period_end: string;
}

export interface DepartmentStat {
	department: string;
	total_employees: number;
	total_present: number;
	total_late: number;
	total_absent: number;
	attendance_rate: number;
}

export interface TrendPoint {
	date: string;
	present: number;
	late: number;
	absent: number;
}

export interface TopLateEmployee {
	employee_id: string;
	name: string;
	department: string;
	late_count: number;
	absent_count: number;
}

export interface ExecutiveSummary {
	summary: string;
	generated_at: string;
}

export interface AnalyticsFilter {
	start_date?: string;
	end_date?: string;
	department_id?: string;
	[key: string]: string | undefined;
}

export const analyticsApi = {
	getAttendanceSummary: (filter: AnalyticsFilter) =>
		api.get<AttendanceSummary>('/api/analytics/attendance-summary', filter),

	getDepartmentStats: (filter: AnalyticsFilter) =>
		api.get<DepartmentStat[]>('/api/analytics/department-stats', filter),

	getTrend: (filter: AnalyticsFilter) =>
		api.get<TrendPoint[]>('/api/analytics/trend', filter),

	getTopLateEmployees: (filter: AnalyticsFilter) =>
		api.get<TopLateEmployee[]>('/api/analytics/top-late-employees', filter),

	getExecutiveSummary: (filter: AnalyticsFilter) =>
		api.get<ExecutiveSummary>('/api/analytics/executive-summary', filter),

	exportPDF: async (filter: AnalyticsFilter) => {
		const blob = await api.download('/api/reports/export-pdf', filter);
		const start = filter.start_date ?? 'start';
		const end = filter.end_date ?? 'end';
		downloadBlob(blob, `laporan-kehadiran-${start}-${end}.pdf`);
	}
};
