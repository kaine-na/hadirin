package document

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository menangani semua operasi database untuk dokumen.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat instance Repository baru.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create menyimpan dokumen baru ke database.
func (r *Repository) Create(ctx context.Context, doc *Document) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO documents (user_id, title, description, category, file_path, file_name, file_size, mime_type, version, parent_id, doc_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`,
		doc.UserID, doc.Title, nullableString(doc.Description), doc.Category,
		doc.FilePath, doc.FileName, doc.FileSize, doc.MimeType, doc.Version,
		nullableString(doc.ParentID), doc.DocDate,
	).Scan(&doc.ID, &doc.CreatedAt)
}

// FindByID mengambil satu dokumen berdasarkan ID.
func (r *Repository) FindByID(ctx context.Context, id string) (*Document, error) {
	doc := &Document{}
	var docDate *time.Time
	var description, parentID string

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id::text, title, COALESCE(description, ''), category,
		       file_path, file_name, file_size, mime_type, version,
		       COALESCE(parent_id::text, ''), doc_date, created_at
		FROM documents WHERE id = $1
	`, id).Scan(
		&doc.ID, &doc.UserID, &doc.Title, &description, &doc.Category,
		&doc.FilePath, &doc.FileName, &doc.FileSize, &doc.MimeType, &doc.Version,
		&parentID, &docDate, &doc.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	doc.Description = description
	doc.ParentID = parentID
	doc.DocDate = docDate
	return doc, nil
}

// List mengambil daftar dokumen dengan filter dan pagination.
func (r *Repository) List(ctx context.Context, filter ListFilter) ([]*Document, int, error) {
	args := []interface{}{}
	argIdx := 1
	conditions := "1=1"

	if filter.UserID != "" {
		conditions += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Category != "" {
		conditions += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, filter.Category)
		argIdx++
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM documents WHERE "+conditions, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count documents: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, user_id::text, title, COALESCE(description, ''), category,
		       file_path, file_name, file_size, mime_type, version,
		       COALESCE(parent_id::text, ''), doc_date, created_at
		FROM documents WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, conditions, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		doc := &Document{}
		var docDate *time.Time
		var description, parentID string
		if err := rows.Scan(
			&doc.ID, &doc.UserID, &doc.Title, &description, &doc.Category,
			&doc.FilePath, &doc.FileName, &doc.FileSize, &doc.MimeType, &doc.Version,
			&parentID, &docDate, &doc.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan document: %w", err)
		}
		doc.Description = description
		doc.ParentID = parentID
		doc.DocDate = docDate
		docs = append(docs, doc)
	}

	return docs, total, nil
}

// Delete menghapus dokumen dari database.
func (r *Repository) Delete(ctx context.Context, id string) (string, error) {
	var filePath string
	err := r.db.QueryRow(ctx, `
		DELETE FROM documents WHERE id = $1 RETURNING file_path
	`, id).Scan(&filePath)
	if err != nil {
		return "", fmt.Errorf("delete document: %w", err)
	}
	return filePath, nil
}

// AddComment menambah komentar pada dokumen.
func (r *Repository) AddComment(ctx context.Context, comment *Comment) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO document_comments (document_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, comment.DocumentID, comment.UserID, comment.Content).Scan(&comment.ID, &comment.CreatedAt)
}

// ListComments mengambil semua komentar untuk satu dokumen.
func (r *Repository) ListComments(ctx context.Context, documentID string) ([]*Comment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, document_id::text, user_id::text, content, created_at
		FROM document_comments
		WHERE document_id = $1
		ORDER BY created_at ASC
	`, documentID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{}
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.UserID, &c.Content, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	return comments, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
