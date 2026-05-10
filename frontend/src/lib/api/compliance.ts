import { api } from './client';

export interface BPJSResult {
	gross_salary: number;
	kes_base_salary: number;
	kes_company: number;
	kes_employee: number;
	kes_total: number;
	kes_company_rate: number;
	kes_employee_rate: number;
	jht_company: number;
	jht_employee: number;
	jht_total: number;
	jht_company_rate: number;
	jht_employee_rate: number;
	jp_base_salary: number;
	jp_company: number;
	jp_employee: number;
	jp_total: number;
	jp_company_rate: number;
	jp_employee_rate: number;
	jkk_company: number;
	jkk_company_rate: number;
	jkm_company: number;
	jkm_company_rate: number;
	total_company_contribution: number;
	total_employee_contribution: number;
	total_contribution: number;
	take_home_pay: number;
}

export interface PPh21Result {
	gross_monthly: number;
	ter_category: 'A' | 'B' | 'C';
	ter_rate: number;
	pph21_amount: number;
	is_december: boolean;
	ytd_gross: number;
	ytd_tax: number;
	annual_gross: number;
	annual_tax: number;
	december_tax: number;
	ptkp: number;
	pkp: number;
}

export interface THRResult {
	base_salary: number;
	service_months: number;
	religion: string;
	holiday_name: string;
	holiday_date?: string;
	deadline_h7?: string;
	days_until_h7: number;
	days_until_holiday: number;
	thr_amount: number;
	is_full_amount: boolean;
	pro_rata_ratio: number;
	is_eligible: boolean;
}

export interface EmployeeTHR {
	user_id: string;
	name: string;
	religion: string;
	thr: THRResult;
}

export interface HolidayInfo {
	religion: string;
	name: string;
	date: string;
	deadline_h7: string;
}

export interface ChecklistItem {
	id: string;
	period: string;
	item_code: string;
	title: string;
	description: string;
	deadline: string;
	status: 'pending' | 'done' | 'overdue';
	done_at?: string;
	done_by?: string;
	notified_h3: boolean;
	days_until: number;
	created_at: string;
	updated_at: string;
}

export interface ChecklistStats {
	pending: number;
	done: number;
	overdue: number;
	total: number;
}

export interface ComplianceSummary {
	period: string;
	overall_status: 'green' | 'yellow' | 'red';
	checklist: {
		items: ChecklistItem[];
		pending: number;
		done: number;
		overdue: number;
	};
	bpjs_summary: {
		total_employees: number;
		total_gross_salary: number;
		total_company_contribution: number;
		total_employee_deduction: number;
	};
	upcoming_deadlines: ChecklistItem[];
}

export const complianceApi = {
	getBPJSCalculation: async (params: {
		month?: string;
		gross_salary: number;
		user_id?: string;
	}): Promise<{ period: string; result: BPJSResult }> => {
		const query = new URLSearchParams({
			gross_salary: String(params.gross_salary)
		});
		if (params.month) query.set('month', params.month);
		if (params.user_id) query.set('user_id', params.user_id);
		return api.get(`/api/compliance/bpjs-calculation?${query}`);
	},

	getPPh21Calculation: async (params: {
		month?: string;
		gross_salary: number;
		marital?: 'TK' | 'K';
		dependents?: number;
		ytd_gross?: number;
		ytd_tax?: number;
	}): Promise<{ period: string; result: PPh21Result }> => {
		const query = new URLSearchParams({
			gross_salary: String(params.gross_salary)
		});
		if (params.month) query.set('month', params.month);
		if (params.marital) query.set('marital', params.marital);
		if (params.dependents !== undefined) query.set('dependents', String(params.dependents));
		if (params.ytd_gross) query.set('ytd_gross', String(params.ytd_gross));
		if (params.ytd_tax) query.set('ytd_tax', String(params.ytd_tax));
		return api.get(`/api/compliance/pph21-calculation?${query}`);
	},

	getTHRCalculation: async (year?: number): Promise<{
		year: number;
		employees: EmployeeTHR[];
		total_thr: number;
		count: number;
		holidays: HolidayInfo[];
	}> => {
		const query = year ? `?year=${year}` : '';
		return api.get(`/api/compliance/thr-calculation${query}`);
	},

	getChecklist: async (month?: string): Promise<{
		period: string;
		items: ChecklistItem[];
		overall_status: 'green' | 'yellow' | 'red';
		stats: ChecklistStats;
	}> => {
		const query = month ? `?month=${month}` : '';
		return api.get(`/api/compliance/checklist${query}`);
	},

	markChecklistDone: async (id: string): Promise<ChecklistItem> => {
		return api.put(`/api/compliance/checklist/${id}/done`, {});
	},

	getSummary: async (month?: string): Promise<ComplianceSummary> => {
		const query = month ? `?month=${month}` : '';
		return api.get(`/api/compliance/summary${query}`);
	}
};
