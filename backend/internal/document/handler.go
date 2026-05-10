package document

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"saas-karyawan/internal/auth"
	"saas-karyawan/pkg/response"
)

// Handler menangani HTTP request untuk modul document.
type Handler struct {
	svc *Service
}

// NewHandler membuat instance Handler baru.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Upload menangani POST /api/documents/upload
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	// Parse multipart form
	maxSize := h.svc.maxFileSizeMB * 1024 * 1024
	if err := r.ParseMultipartForm(maxSize); err != nil {
		response.Error(w, http.StatusBadRequest, "gagal parse form atau file terlalu besar")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "file tidak ditemukan dalam request")
		return
	}
	defer file.Close()

	req := &UploadRequest{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Category:    r.FormValue("category"),
		DocDate:     r.FormValue("doc_date"),
	}

	doc, err := h.svc.Upload(r.Context(), claims.UserID, req, file, header)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(w, "dokumen berhasil diupload", doc)
}

// List menangani GET /api/documents
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	filter := ListFilter{
		Category: r.URL.Query().Get("category"),
	}

	// HR/Manager bisa lihat semua, karyawan hanya milik sendiri
	userIDParam := r.URL.Query().Get("user_id")
	if claims.Role == "hr_admin" || claims.Role == "super_admin" || claims.Role == "manager" {
		filter.UserID = userIDParam // Bisa filter by user_id atau kosong (semua)
	} else {
		filter.UserID = claims.UserID // Karyawan hanya lihat milik sendiri
	}

	docs, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil daftar dokumen")
		return
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	response.Paginated(w, "daftar dokumen berhasil diambil", docs, total, filter.Page, filter.PageSize)
}

// GetByID menangani GET /api/documents/:id
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	doc, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "dokumen tidak ditemukan")
		return
	}

	response.Success(w, "detail dokumen berhasil diambil", doc)
}

// Delete menangani DELETE /api/documents/:id
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetClaims(r)

	if err := h.svc.Delete(r.Context(), id, claims.UserID, claims.Role); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	response.Success(w, "dokumen berhasil dihapus", nil)
}

// Download menangani GET /api/documents/:id/download
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	doc, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "dokumen tidak ditemukan")
		return
	}

	// Buka file dari disk
	f, err := os.Open(doc.FilePath)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "file tidak dapat dibuka")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", doc.MimeType)
	// RFC 6266: quote filename untuk menangani spasi dan karakter khusus
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.FileName))
	http.ServeContent(w, r, doc.FileName, doc.CreatedAt, f)
}

// AddComment menangani POST /api/documents/:id/comments
func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetClaims(r)

	var req AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "request body tidak valid")
		return
	}

	comment, err := h.svc.AddComment(r.Context(), id, claims.UserID, req.Content)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(w, "komentar berhasil ditambahkan", comment)
}

// ListComments menangani GET /api/documents/:id/comments
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	comments, err := h.svc.ListComments(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil komentar")
		return
	}

	response.Success(w, "komentar berhasil diambil", comments)
}
