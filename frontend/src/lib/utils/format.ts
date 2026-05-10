/**
 * Format tanggal ke string Bahasa Indonesia.
 * Contoh: "10 Mei 2026"
 */
export function formatDate(input: string | Date | null | undefined, withYear = true): string {
	if (!input) return '-';
	const date = typeof input === 'string' ? new Date(input) : input;
	if (Number.isNaN(date.getTime())) return '-';
	const opts: Intl.DateTimeFormatOptions = {
		day: 'numeric',
		month: 'long',
		year: withYear ? 'numeric' : undefined
	};
	return date.toLocaleDateString('id-ID', opts);
}

export function formatDateTime(input: string | Date | null | undefined): string {
	if (!input) return '-';
	const date = typeof input === 'string' ? new Date(input) : input;
	if (Number.isNaN(date.getTime())) return '-';
	return date.toLocaleString('id-ID', {
		day: 'numeric',
		month: 'long',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit'
	});
}

export function formatTime(input: string | Date | null | undefined): string {
	if (!input) return '-';
	const date = typeof input === 'string' ? new Date(input) : input;
	if (Number.isNaN(date.getTime())) return '-';
	return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
}

export function formatFileSize(bytes: number): string {
	if (!bytes || bytes < 0) return '0 B';
	const units = ['B', 'KB', 'MB', 'GB'];
	let size = bytes;
	let i = 0;
	while (size >= 1024 && i < units.length - 1) {
		size /= 1024;
		i++;
	}
	return `${size.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

export function todayISO(): string {
	const d = new Date();
	const year = d.getFullYear();
	const month = String(d.getMonth() + 1).padStart(2, '0');
	const day = String(d.getDate()).padStart(2, '0');
	return `${year}-${month}-${day}`;
}

export function startOfMonthISO(): string {
	const d = new Date();
	const year = d.getFullYear();
	const month = String(d.getMonth() + 1).padStart(2, '0');
	return `${year}-${month}-01`;
}

export function roleLabel(role: string): string {
	const map: Record<string, string> = {
		super_admin: 'Super Admin',
		hr_admin: 'HR Admin',
		manager: 'Manager',
		karyawan: 'Karyawan'
	};
	return map[role] ?? role;
}

export function statusLabel(status: string): string {
	const map: Record<string, string> = {
		hadir: 'Hadir',
		terlambat: 'Terlambat',
		izin: 'Izin',
		sakit: 'Sakit',
		alpha: 'Alpha'
	};
	return map[status] ?? status;
}

/**
 * Warna badge berdasarkan status absensi.
 */
export function statusColor(status: string): 'green' | 'yellow' | 'blue' | 'red' | 'slate' {
	switch (status) {
		case 'hadir':
			return 'green';
		case 'terlambat':
			return 'yellow';
		case 'izin':
			return 'blue';
		case 'sakit':
			return 'blue';
		case 'alpha':
			return 'red';
		default:
			return 'slate';
	}
}

/**
 * Format angka ke format mata uang Rupiah.
 * Contoh: 5000000 → "Rp 5.000.000"
 */
export function formatCurrency(amount: number | null | undefined): string {
	if (amount == null) return 'Rp 0';
	return new Intl.NumberFormat('id-ID', {
		style: 'currency',
		currency: 'IDR',
		minimumFractionDigits: 0,
		maximumFractionDigits: 0
	}).format(amount);
}
