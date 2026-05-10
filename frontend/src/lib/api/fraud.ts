import { api } from './client';

export interface FraudLog {
	id: string;
	attendance_id?: string;
	user_id: string;
	employee_name?: string;
	fraud_type: string;
	severity: 'low' | 'medium' | 'high' | 'critical';
	description: string;
	evidence?: Record<string, unknown>;
	status: 'pending' | 'dismissed' | 'confirmed';
	reviewed_by?: string;
	reviewed_at?: string;
	review_notes?: string;
	ai_analysis?: string;
	ai_confidence?: number;
	photo_url?: string;
	created_at: string;
	updated_at: string;
}

export interface FraudSummary {
	total_logs: number;
	pending_logs: number;
	confirmed_logs: number;
	dismissed_logs: number;
	by_type: Record<string, number>;
	by_severity: Record<string, number>;
	top_employees: Array<{
		user_id: string;
		employee_name: string;
		fraud_count: number;
	}>;
}

export interface FraudValidationResult {
	fraud_detected: boolean;
	fraud_count: number;
	fraud_results: Array<{
		type: string;
		severity: string;
		description: string;
		evidence?: Record<string, unknown>;
		ai_analysis?: string;
	}>;
	liveness?: {
		is_live_face: boolean;
		score: number;
		notes: string;
		photo_id?: string;
	};
}

export interface PaginatedFraudLogs {
	items: FraudLog[];
	total: number;
	page: number;
	page_size: number;
}

export const fraudApi = {
	// Validasi clock-in dengan GPS + selfie
	async validateClockIn(
		attendanceId: string,
		gps: { latitude: number; longitude: number; accuracy: number },
		selfie?: File
	): Promise<FraudValidationResult> {
		const form = new FormData();
		form.append('attendance_id', attendanceId);
		form.append('latitude', String(gps.latitude));
		form.append('longitude', String(gps.longitude));
		form.append('accuracy', String(gps.accuracy));
		if (selfie) {
			form.append('selfie', selfie);
		}
		return api.upload<FraudValidationResult>('/api/fraud/validate-clock-in', form);
	},

	// List fraud logs (HR only)
	async listLogs(params?: {
		status?: string;
		page?: number;
		page_size?: number;
	}): Promise<PaginatedFraudLogs> {
		return api.get<PaginatedFraudLogs>('/api/fraud/logs', {
			status: params?.status,
			page: params?.page,
			page_size: params?.page_size
		});
	},

	// Detail fraud log
	async getLog(id: string): Promise<FraudLog> {
		return api.get<FraudLog>(`/api/fraud/logs/${id}`);
	},

	// Dismiss fraud log (false positive)
	async dismissLog(id: string, notes?: string): Promise<void> {
		return api.put<void>(`/api/fraud/logs/${id}/dismiss`, { notes: notes ?? '' });
	},

	// Konfirmasi fraud log
	async confirmLog(id: string, notes?: string): Promise<void> {
		return api.put<void>(`/api/fraud/logs/${id}/confirm`, { notes: notes ?? '' });
	},

	// Ringkasan fraud bulan ini
	async getSummary(): Promise<FraudSummary> {
		return api.get<FraudSummary>('/api/fraud/summary');
	}
};
