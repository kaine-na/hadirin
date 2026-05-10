package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service menangani logika bisnis untuk modul document.
type Service struct {
	repo          *Repository
	uploadDir     string
	maxFileSizeMB int64
}

// NewService membuat instance Service baru.
func NewService(db *pgxpool.Pool, uploadDir string, maxFileSizeMB int64) *Service {
	return &Service{
		repo:          NewRepository(db),
		uploadDir:     uploadDir,
		maxFileSizeMB: maxFileSizeMB,
	}
}

// Upload memproses upload dokumen baru.
func (s *Service) Upload(ctx context.Context, userID string, req *UploadRequest, file multipart.File, header *multipart.FileHeader) (*Document, error) {
	// Validasi input
	if req.Title == "" {
		return nil, errors.New("judul dokumen wajib diisi")
	}
	if req.Category == "" {
		return nil, errors.New("kategori dokumen wajib diisi")
	}

	// Validasi ukuran file
	maxSize := s.maxFileSizeMB * 1024 * 1024
	if header.Size > maxSize {
		return nil, fmt.Errorf("ukuran file melebihi batas %dMB", s.maxFileSizeMB)
	}

	// Validasi MIME type via magic bytes (bukan hanya ekstensi)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("baca file: %w", err)
	}
	mimeType := http.DetectContentType(buf[:n])

	// Kembalikan pointer ke awal file setelah baca magic bytes
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek file: %w", err)
	}

	// Whitelist MIME types yang diizinkan
	allowedMimes := map[string]bool{
		"application/pdf":                                                    true,
		"image/jpeg":                                                         true,
		"image/png":                                                          true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/msword":                                                 true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"application/vnd.ms-excel":                                          true,
		"text/plain":                                                         true,
	}
	if !allowedMimes[mimeType] {
		return nil, fmt.Errorf("tipe file tidak diizinkan: %s", mimeType)
	}

	// Buat direktori upload jika belum ada
	docDir := filepath.Join(s.uploadDir, "documents", userID)
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	// Sanitasi nama file dan generate nama unik
	safeFilename := sanitizeFilename(header.Filename)
	ext := filepath.Ext(safeFilename)
	uniqueFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(safeFilename, ext), ext)
	filePath := filepath.Join(docDir, uniqueFilename)

	// Simpan file ke disk
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	// Parse doc_date jika ada
	var docDate *time.Time
	if req.DocDate != "" {
		t, err := time.Parse("2006-01-02", req.DocDate)
		if err == nil {
			docDate = &t
		}
	}

	doc := &Document{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		FilePath:    filePath,
		FileName:    header.Filename,
		FileSize:    header.Size,
		MimeType:    mimeType,
		Version:     1,
		DocDate:     docDate,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		// Hapus file jika gagal simpan ke DB
		_ = os.Remove(filePath)
		return nil, fmt.Errorf("save document: %w", err)
	}

	return doc, nil
}

// GetByID mengambil detail satu dokumen.
func (s *Service) GetByID(ctx context.Context, id string) (*Document, error) {
	return s.repo.FindByID(ctx, id)
}

// List mengambil daftar dokumen dengan filter.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]*Document, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return s.repo.List(ctx, filter)
}

// Delete menghapus dokumen dan file-nya.
func (s *Service) Delete(ctx context.Context, id, requestingUserID, requestingRole string) error {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("dokumen tidak ditemukan")
	}

	// Hanya owner atau HR yang boleh hapus
	if doc.UserID != requestingUserID && requestingRole != "hr_admin" && requestingRole != "super_admin" {
		return errors.New("tidak memiliki izin untuk menghapus dokumen ini")
	}

	filePath, err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	// Hapus file dari disk (abaikan error jika file sudah tidak ada)
	_ = os.Remove(filePath)
	return nil
}

// AddComment menambah komentar pada dokumen.
func (s *Service) AddComment(ctx context.Context, documentID, userID, content string) (*Comment, error) {
	if content == "" {
		return nil, errors.New("konten komentar tidak boleh kosong")
	}

	// Pastikan dokumen ada
	if _, err := s.repo.FindByID(ctx, documentID); err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}

	comment := &Comment{
		DocumentID: documentID,
		UserID:     userID,
		Content:    content,
	}

	if err := s.repo.AddComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("add comment: %w", err)
	}

	return comment, nil
}

// ListComments mengambil semua komentar untuk satu dokumen.
func (s *Service) ListComments(ctx context.Context, documentID string) ([]*Comment, error) {
	return s.repo.ListComments(ctx, documentID)
}

// sanitizeFilename membersihkan nama file dari karakter berbahaya.
func sanitizeFilename(name string) string {
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			result.WriteRune(r)
		}
	}
	s := result.String()
	if s == "" {
		return "file"
	}
	return s
}
