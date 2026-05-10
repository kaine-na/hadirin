// Types untuk komponen shared
export interface Column<Row> {
	key: string;
	label: string;
	field?: keyof Row;
	align?: 'left' | 'center' | 'right';
	width?: string;
}
