import { api } from './client';
import type { DocumentComment, DocumentItem, Paginated } from '$lib/types';

export interface DocumentListQuery {
	user_id?: string;
	category?: string;
	page?: number;
	page_size?: number;
	[key: string]: string | number | boolean | undefined | null;
}

export interface UploadDocumentPayload {
	title: string;
	description: string;
	category: string;
	doc_date: string;
	file: File;
}

export const documentsApi = {
	list: (q: DocumentListQuery = {}) => api.get<Paginated<DocumentItem>>('/api/documents', q),
	get: (id: string) => api.get<DocumentItem>(`/api/documents/${id}`),
	upload: (payload: UploadDocumentPayload) => {
		const form = new FormData();
		form.append('title', payload.title);
		form.append('description', payload.description);
		form.append('category', payload.category);
		form.append('doc_date', payload.doc_date);
		form.append('file', payload.file);
		return api.upload<DocumentItem>('/api/documents/upload', form);
	},
	remove: (id: string) => api.del<null>(`/api/documents/${id}`),
	download: (id: string) => api.download(`/api/documents/${id}/download`),
	listComments: (id: string) => api.get<DocumentComment[]>(`/api/documents/${id}/comments`),
	addComment: (id: string, content: string) =>
		api.post<DocumentComment>(`/api/documents/${id}/comments`, { content })
};
