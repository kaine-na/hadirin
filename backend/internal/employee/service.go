package employee

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
	"golang.org/x/crypto/bcrypt"
)

// Service menangani logika bisnis untuk modul employee.
type Service struct {
	repo      *Repository
	uploadDir string
}

// NewService membuat instance Service baru.
func NewService(db *pgxpool.Pool, uploadDir string) *Service {
	return &Service{
		repo:      NewRepository(db),
		uploadDir: uploadDir,
	}
}

// List mengambil daftar karyawan dengan filter dan pagination.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]*Employee, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return s.repo.List(ctx, filter)
}

// GetByID mengambil detail satu karyawan.
func (s *Service) GetByID(ctx context.Context, id string) (*Employee, error) {
	return s.repo.FindByID(ctx, id)
}

// Create membuat karyawan baru.
func (s *Service) Create(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error) {
	// Validasi input
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New("nama, email, dan password wajib diisi")
	}

	// Cek email duplikat
	if existing, _ := s.repo.FindByEmail(ctx, req.Email); existing != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	// Default role
	if req.Role == "" {
		req.Role = "karyawan"
	}

	// Validasi role
	validRoles := map[string]bool{"super_admin": true, "hr_admin": true, "manager": true, "karyawan": true}
	if !validRoles[req.Role] {
		return nil, fmt.Errorf("role tidak valid: %s", req.Role)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	emp := &Employee{
		Name:       req.Name,
		Email:      req.Email,
		Role:       req.Role,
		Department: req.Department,
		Position:   req.Position,
		NIK:        req.NIK,
	}

	if err := s.repo.Create(ctx, emp, string(hash)); err != nil {
		return nil, fmt.Errorf("create employee: %w", err)
	}

	return emp, nil
}

// Update mengupdate data karyawan.
func (s *Service) Update(ctx context.Context, id string, req *UpdateEmployeeRequest) (*Employee, error) {
	// Pastikan karyawan ada
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	return s.repo.Update(ctx, id, req)
}

// Delete menonaktifkan karyawan (soft delete).
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}

// UploadPhoto mengupload foto profil karyawan.
func (s *Service) UploadPhoto(ctx context.Context, employeeID string, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validasi tipe file via magic bytes (bukan header Content-Type yang bisa di-spoof)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("baca file: %w", err)
	}
	detectedMime := http.DetectContentType(buf[:n])

	// Kembalikan pointer ke awal file setelah baca magic bytes
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek file: %w", err)
	}

	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowedTypes[detectedMime] {
		return "", errors.New("tipe file tidak didukung, gunakan JPEG/PNG/WebP")
	}

	// Buat direktori photos jika belum ada
	photoDir := filepath.Join(s.uploadDir, "photos")
	if err := os.MkdirAll(photoDir, 0755); err != nil {
		return "", fmt.Errorf("create photo dir: %w", err)
	}

	// Generate nama file unik dengan ekstensi yang disanitasi
	safeFilename := sanitizeFilename(header.Filename)
	ext := filepath.Ext(safeFilename)
	// Pastikan ekstensi sesuai dengan MIME type yang terdeteksi
	filename := fmt.Sprintf("%s_%d%s", employeeID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(photoDir, filename)

	// Simpan file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}

	// URL relatif untuk disimpan di database
	photoURL := "/uploads/photos/" + filename

	// Update di database
	if err := s.repo.UpdatePhoto(ctx, employeeID, photoURL); err != nil {
		return "", fmt.Errorf("update photo url: %w", err)
	}

	return photoURL, nil
}

// sanitizeFilename membersihkan nama file dari karakter berbahaya.
func sanitizeFilename(name string) string {
	// Hanya izinkan alphanumeric, dash, underscore, dan titik
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
