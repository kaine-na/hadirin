package document

import "time"

// Document adalah representasi data dokumen dari database.
type Document struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Category    string     `json:"category"`
	FilePath    string     `json:"-"` // Tidak dikirim ke client
	FileName    string     `json:"file_name"`
	FileSize    int64      `json:"file_size"`
	MimeType    string     `json:"mime_type"`
	Version     int        `json:"version"`
	ParentID    string     `json:"parent_id,omitempty"`
	DocDate     *time.Time `json:"doc_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Comment adalah komentar HR/Manager pada dokumen.
type Comment struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	UserID     string    `json:"user_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

// UploadRequest adalah metadata yang dikirim bersama file upload.
type UploadRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	DocDate     string `json:"doc_date"` // Format: YYYY-MM-DD
}

// AddCommentRequest adalah body request untuk menambah komentar.
type AddCommentRequest struct {
	Content string `json:"content"`
}

// ListFilter adalah parameter filter untuk list dokumen.
type ListFilter struct {
	UserID   string
	Category string
	Page     int
	PageSize int
}
