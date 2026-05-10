// TypeScript types untuk SaaS Karyawan

export type Role = 'super_admin' | 'hr_admin' | 'manager' | 'karyawan';

export type AttendanceStatus = 'hadir' | 'terlambat' | 'izin' | 'sakit' | 'alpha';

export type DocumentCategory =
	| 'Laporan Harian'
	| 'Laporan Mingguan'
	| 'Laporan Proyek'
	| 'Lainnya';

export interface User {
	id: string;
	name: string;
	email: string;
	role: Role;
	department?: string;
	position?: string;
	photo_url?: string;
}

export interface Employee {
	id: string;
	company_id: string;
	name: string;
	email: string;
	role: Role;
	department?: string;
	position?: string;
	nik?: string;
	photo_url?: string;
	joined_at?: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface CreateEmployeePayload {
	name: string;
	email: string;
	password: string;
	role: Role;
	department: string;
	position: string;
	nik: string;
}

export interface UpdateEmployeePayload {
	name: string;
	role: Role;
	department: string;
	position: string;
	nik: string;
	is_active?: boolean;
}

export interface Attendance {
	id: string;
	user_id: string;
	date: string;
	clock_in?: string;
	clock_out?: string;
	status: AttendanceStatus;
	notes?: string;
	ip_address?: string;
	created_by?: string;
	updated_by?: string;
	created_at: string;
	updated_at: string;
}

export interface AttendanceOverridePayload {
	status: AttendanceStatus;
	notes: string;
	clock_in: string;
	clock_out: string;
}

export interface DocumentItem {
	id: string;
	user_id: string;
	title: string;
	description?: string;
	category: string;
	file_name: string;
	file_size: number;
	mime_type: string;
	version: number;
	parent_id?: string;
	doc_date?: string;
	created_at: string;
}

export interface DocumentComment {
	id: string;
	document_id: string;
	user_id: string;
	content: string;
	created_at: string;
}

export interface AIReport {
	id: string;
	employee_id: string;
	generated_by: string;
	period_start: string;
	period_end: string;
	prompt?: string;
	response: string;
	model_used?: string;
	created_at: string;
}

export interface AnalyzePayload {
	period_start: string;
	period_end: string;
	custom_prompt?: string;
}

export interface ApiResponse<T = unknown> {
	success: boolean;
	message: string;
	data?: T;
}

export interface Paginated<T> {
	items: T[];
	total: number;
	page: number;
	page_size: number;
	total_pages: number;
}

export interface LoginPayload {
	email: string;
	password: string;
}

export interface LoginResponse {
	token: string;
	expires_at: string;
	user: User;
}
