package ai

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"saas-karyawan/internal/auth"
	"saas-karyawan/pkg/response"
)

// Handler menangani HTTP request untuk modul AI.
type Handler struct {
	svc *Service
}

// NewHandler membuat instance Handler baru.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Analyze menangani POST /api/ai/analyze/:employee_id
func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employee_id")
	claims := auth.GetClaims(r)

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "request body tidak valid")
		return
	}

	if req.PeriodStart == "" || req.PeriodEnd == "" {
		response.Error(w, http.StatusBadRequest, "period_start dan period_end wajib diisi")
		return
	}

	// Validasi format dan urutan tanggal
	start, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "format period_start tidak valid, gunakan YYYY-MM-DD")
		return
	}
	end, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "format period_end tidak valid, gunakan YYYY-MM-DD")
		return
	}
	if start.After(end) {
		response.Error(w, http.StatusBadRequest, "period_start tidak boleh setelah period_end")
		return
	}

	report, err := h.svc.Analyze(r.Context(), employeeID, claims.UserID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(w, "laporan AI berhasil dibuat", report)
}

// GetReports menangani GET /api/ai/reports/:employee_id
func (h *Handler) GetReports(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employee_id")

	reports, err := h.svc.GetReports(r.Context(), employeeID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil riwayat laporan")
		return
	}

	response.Success(w, "riwayat laporan AI berhasil diambil", reports)
}

// GetReportByID menangani GET /api/ai/reports/:id
func (h *Handler) GetReportByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	report, err := h.svc.GetReportByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "laporan tidak ditemukan")
		return
	}

	response.Success(w, "detail laporan AI berhasil diambil", report)
}
