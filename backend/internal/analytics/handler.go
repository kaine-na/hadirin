package analytics

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler menangani HTTP request untuk modul analytics.
type Handler struct {
	svc *Service
}

// NewHandler membuat instance Handler baru.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// parseFilter mengambil filter dari query params.
func parseFilter(r *http.Request) AnalyticsFilter {
	q := r.URL.Query()

	startDate := q.Get("start_date")
	endDate := q.Get("end_date")

	// Default: 30 hari terakhir
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	return AnalyticsFilter{
		StartDate:    startDate,
		EndDate:      endDate,
		DepartmentID: q.Get("department_id"),
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GetAttendanceSummary menangani GET /api/analytics/attendance-summary
func (h *Handler) GetAttendanceSummary(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	summary, err := h.svc.GetAttendanceSummary(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil ringkasan kehadiran: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// GetDepartmentStats menangani GET /api/analytics/department-stats
func (h *Handler) GetDepartmentStats(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	stats, err := h.svc.GetDepartmentStats(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil statistik departemen: "+err.Error())
		return
	}
	if stats == nil {
		stats = []*DepartmentStat{}
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetTrend menangani GET /api/analytics/trend
func (h *Handler) GetTrend(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	trend, err := h.svc.GetTrend(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil tren kehadiran: "+err.Error())
		return
	}
	if trend == nil {
		trend = []*TrendPoint{}
	}
	writeJSON(w, http.StatusOK, trend)
}

// GetTopLateEmployees menangani GET /api/analytics/top-late-employees
func (h *Handler) GetTopLateEmployees(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	employees, err := h.svc.GetTopLateEmployees(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data karyawan terlambat: "+err.Error())
		return
	}
	if employees == nil {
		employees = []*TopLateEmployee{}
	}
	writeJSON(w, http.StatusOK, employees)
}

// GetExecutiveSummary menangani GET /api/analytics/executive-summary
func (h *Handler) GetExecutiveSummary(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	result, err := h.svc.GenerateExecutiveSummary(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal generate executive summary: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ExportPDF menangani GET /api/reports/export-pdf
func (h *Handler) ExportPDF(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)

	summary, err := h.svc.GetAttendanceSummary(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
		return
	}

	deptStats, err := h.svc.GetDepartmentStats(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil statistik departemen: "+err.Error())
		return
	}

	topLate, err := h.svc.GetTopLateEmployees(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data karyawan: "+err.Error())
		return
	}

	pdfBytes, err := GeneratePDF(summary, deptStats, topLate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal generate PDF: "+err.Error())
		return
	}

	filename := "laporan-kehadiran-" + f.StartDate + "-" + f.EndDate + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes) //nolint:errcheck
}
