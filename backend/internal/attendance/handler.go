package attendance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"saas-karyawan/internal/auth"
	"saas-karyawan/pkg/response"
)

// Handler menangani HTTP request untuk modul attendance.
type Handler struct {
	svc *Service
}

// NewHandler membuat instance Handler baru.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ClockIn menangani POST /api/attendance/clock-in
func (h *Handler) ClockIn(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	var req ClockInRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Ambil IP address dari request
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	att, err := h.svc.ClockIn(r.Context(), claims.UserID, ipAddress, req.Notes)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(w, "clock-in berhasil", att)
}

// ClockOut menangani POST /api/attendance/clock-out
func (h *Handler) ClockOut(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	var req ClockOutRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	att, err := h.svc.ClockOut(r.Context(), claims.UserID, req.Notes)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, "clock-out berhasil", att)
}

// GetToday menangani GET /api/attendance/today
func (h *Handler) GetToday(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	att, err := h.svc.GetToday(r.Context(), claims.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil data absensi")
		return
	}

	response.Success(w, "data absensi hari ini", att)
}

// GetMe menangani GET /api/attendance/me
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	filter := RekapFilter{
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
	}

	attendances, total, err := h.svc.GetByEmployee(r.Context(), claims.UserID, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil rekap absensi")
		return
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 31
	}

	response.Paginated(w, "rekap absensi berhasil diambil", attendances, total, filter.Page, filter.PageSize)
}

// GetByEmployee menangani GET /api/attendance/:employee_id (HR only)
func (h *Handler) GetByEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employee_id")

	filter := RekapFilter{
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
	}

	attendances, total, err := h.svc.GetByEmployee(r.Context(), employeeID, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil rekap absensi")
		return
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 31
	}

	response.Paginated(w, "rekap absensi karyawan berhasil diambil", attendances, total, filter.Page, filter.PageSize)
}

// Override menangani PUT /api/attendance/:id (HR only)
func (h *Handler) Override(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetClaims(r)

	var req OverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "request body tidak valid")
		return
	}

	att, err := h.svc.Override(r.Context(), id, claims.UserID, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, "absensi berhasil diupdate", att)
}

// ExportCSV menangani GET /api/attendance/export/csv (HR only)
func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Default: bulan ini
	if startDate == "" {
		now := time.Now()
		startDate = fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	csvData, err := h.svc.ExportCSV(r.Context(), startDate, endDate)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal export CSV")
		return
	}

	filename := fmt.Sprintf("absensi_%s_%s.csv", startDate, endDate)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}
