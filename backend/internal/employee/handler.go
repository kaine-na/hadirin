package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"saas-karyawan/internal/auth"
	"saas-karyawan/pkg/response"
)

// Handler menangani HTTP request untuk modul employee.
type Handler struct {
	svc *Service
}

// NewHandler membuat instance Handler baru.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List menangani GET /api/employees
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := ListFilter{
		Department: r.URL.Query().Get("department"),
		Role:       r.URL.Query().Get("role"),
		Search:     r.URL.Query().Get("search"),
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	filter.Page = page
	filter.PageSize = pageSize

	employees, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil data karyawan")
		return
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	response.Paginated(w, "data karyawan berhasil diambil", employees, total, filter.Page, filter.PageSize)
}

// GetByID menangani GET /api/employees/:id
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	emp, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "karyawan tidak ditemukan")
		return
	}
	response.Success(w, "data karyawan berhasil diambil", emp)
}

// Create menangani POST /api/employees
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "request body tidak valid")
		return
	}

	emp, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(w, "karyawan berhasil dibuat", emp)
}

// Update menangani PUT /api/employees/:id
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "request body tidak valid")
		return
	}

	emp, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, "data karyawan berhasil diupdate", emp)
}

// Delete menangani DELETE /api/employees/:id
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, "karyawan berhasil dinonaktifkan", nil)
}

// UploadPhoto menangani POST /api/employees/:id/photo
func (h *Handler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Cek apakah user boleh upload foto ini (diri sendiri atau HR)
	claims := auth.GetClaims(r)
	if claims != nil && claims.UserID != id && claims.Role != "hr_admin" && claims.Role != "super_admin" {
		response.Error(w, http.StatusForbidden, "tidak boleh mengupload foto karyawan lain")
		return
	}

	// Parse multipart form (max 5MB untuk foto)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "file terlalu besar (max 5MB)")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "file photo tidak ditemukan")
		return
	}
	defer file.Close()

	photoURL, err := h.svc.UploadPhoto(r.Context(), id, file, header)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, "foto berhasil diupload", map[string]string{"photo_url": photoURL})
}
