package compliance

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"saas-karyawan/internal/auth"
	"saas-karyawan/internal/notification"
	"saas-karyawan/pkg/response"
)

// Handler menangani HTTP request untuk modul compliance.
type Handler struct {
	db               *pgxpool.Pool
	checklistRepo    *ChecklistRepository
	notificationSvc  *notification.Service
}

// NewHandler membuat instance Handler baru.
func NewHandler(db *pgxpool.Pool, notifSvc *notification.Service) *Handler {
	return &Handler{
		db:              db,
		checklistRepo:   NewChecklistRepository(db),
		notificationSvc: notifSvc,
	}
}

// GetBPJSCalculation menghitung BPJS untuk karyawan tertentu.
// GET /api/compliance/bpjs-calculation?month=2026-05&user_id=xxx
func (h *Handler) GetBPJSCalculation(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	grossSalaryStr := r.URL.Query().Get("gross_salary")
	if grossSalaryStr == "" {
		// Ambil dari database jika user_id diberikan
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			// Gunakan user yang sedang login
			claims := auth.GetClaims(r)
			if claims != nil {
				userID = claims.UserID
			}
		}

		if userID != "" {
			salary, err := h.getUserSalary(r.Context(), userID)
			if err == nil {
				grossSalaryStr = strconv.FormatInt(salary, 10)
			}
		}
	}

	grossSalary, err := strconv.ParseInt(grossSalaryStr, 10, 64)
	if err != nil || grossSalary <= 0 {
		response.Error(w, http.StatusBadRequest, "gross_salary tidak valid")
		return
	}

	period, err := time.Parse("2006-01", month)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "format month tidak valid, gunakan YYYY-MM")
		return
	}

	result := CalculateBPJSForPeriod(grossSalary, period)
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"period": month,
		"result": result,
	})
}

// GetPPh21Calculation menghitung PPh 21 untuk karyawan tertentu.
// GET /api/compliance/pph21-calculation?month=2026-05&user_id=xxx
func (h *Handler) GetPPh21Calculation(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	period, err := time.Parse("2006-01", month)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "format month tidak valid, gunakan YYYY-MM")
		return
	}

	grossSalaryStr := r.URL.Query().Get("gross_salary")
	grossSalary, err := strconv.ParseInt(grossSalaryStr, 10, 64)
	if err != nil || grossSalary <= 0 {
		response.Error(w, http.StatusBadRequest, "gross_salary tidak valid")
		return
	}

	// Parse status PTKP dari query params
	maritalStr := r.URL.Query().Get("marital") // TK atau K
	depsStr := r.URL.Query().Get("dependents")  // 0-3

	marital := MaritalTK
	if maritalStr == "K" {
		marital = MaritalK
	}

	deps, _ := strconv.Atoi(depsStr)
	if deps < 0 {
		deps = 0
	}
	if deps > 3 {
		deps = 3
	}

	ytdGross, _ := strconv.ParseInt(r.URL.Query().Get("ytd_gross"), 10, 64)
	ytdTax, _ := strconv.ParseInt(r.URL.Query().Get("ytd_tax"), 10, 64)

	status := PTKPStatus{
		Marital:    marital,
		Dependents: deps,
	}

	result := CalculatePPh21ForPeriod(grossSalary, status, period, ytdGross, ytdTax)
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"period": month,
		"result": result,
	})
}

// GetTHRCalculation menghitung THR untuk semua karyawan di tahun tertentu.
// GET /api/compliance/thr-calculation?year=2026
func (h *Handler) GetTHRCalculation(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	year := time.Now().Year()
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	// Ambil semua karyawan aktif dari database
	// Catatan: religion dan salary belum ada di schema dasar, gunakan default
	rows, err := h.db.Query(r.Context(), `
		SELECT u.id, u.name,
			COALESCE(u.joined_at, NOW()) as joined_at
		FROM users u
		WHERE u.is_active = true AND u.role != 'super_admin'
		ORDER BY u.name
	`)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil data karyawan")
		return
	}
	defer rows.Close()

	type EmployeeTHR struct {
		UserID   string    `json:"user_id"`
		Name     string    `json:"name"`
		Religion Religion  `json:"religion"`
		THR      THRResult `json:"thr"`
	}

	var results []EmployeeTHR
	for rows.Next() {
		var userID, name string
		var joinedAt time.Time

		if err := rows.Scan(&userID, &name, &joinedAt); err != nil {
			continue
		}

		// Hitung masa kerja dari joined_at
		serviceMonths := int(time.Since(joinedAt).Hours() / 24 / 30)

		// Default: agama Islam, gaji pokok 5 juta (placeholder)
		// Dalam implementasi nyata, ambil dari tabel profil karyawan
		religion := ReligionIslam
		baseSalary := int64(5_000_000)

		thrResult := CalculateTHR(THRInput{
			BaseSalary:    baseSalary,
			ServiceMonths: serviceMonths,
			Religion:      religion,
			Year:          year,
		})

		results = append(results, EmployeeTHR{
			UserID:   userID,
			Name:     name,
			Religion: religion,
			THR:      thrResult,
		})
	}

	if err := rows.Err(); err != nil {
		response.Error(w, http.StatusInternalServerError, "error membaca data karyawan")
		return
	}

	// Hitung total THR
	var totalTHR int64
	for _, r := range results {
		totalTHR += r.THR.THRAmount
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"year":       year,
		"employees":  results,
		"total_thr":  totalTHR,
		"count":      len(results),
		"holidays":   GetHolidays(year),
	})
}

// GetChecklist mengambil checklist kepatuhan untuk bulan tertentu.
// GET /api/compliance/checklist?month=2026-05
func (h *Handler) GetChecklist(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	// Validasi format
	if _, err := time.Parse("2006-01", month); err != nil {
		response.Error(w, http.StatusBadRequest, "format month tidak valid, gunakan YYYY-MM")
		return
	}

	// Update overdue items terlebih dahulu
	_, _ = h.checklistRepo.UpdateOverdueItems(r.Context())

	// Generate checklist jika belum ada, lalu ambil
	items, err := h.checklistRepo.GenerateMonthlyChecklist(r.Context(), month)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("gagal generate checklist: %v", err))
		return
	}

	// Hitung statistik
	var pending, done, overdue int
	for _, item := range items {
		switch item.Status {
		case "done":
			done++
		case "overdue":
			overdue++
		default:
			pending++
		}
	}

	// Tentukan status keseluruhan
	overallStatus := "green"
	if overdue > 0 {
		overallStatus = "red"
	} else {
		// Cek apakah ada yang mendekati deadline (dalam 3 hari)
		for _, item := range items {
			if item.Status == "pending" && item.DaysUntil <= 3 {
				overallStatus = "yellow"
				break
			}
		}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"period":         month,
		"items":          items,
		"overall_status": overallStatus,
		"stats": map[string]int{
			"pending": pending,
			"done":    done,
			"overdue": overdue,
			"total":   len(items),
		},
	})
}

// MarkChecklistDone menandai item checklist sebagai selesai.
// PUT /api/compliance/checklist/:id/done
func (h *Handler) MarkChecklistDone(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "id wajib diisi")
		return
	}

	claims := auth.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	item, err := h.checklistRepo.MarkDone(r.Context(), id, claims.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("gagal menandai selesai: %v", err))
		return
	}

	response.JSON(w, http.StatusOK, item)
}

// GetSummary mengembalikan ringkasan semua kewajiban compliance bulan ini.
// GET /api/compliance/summary?month=2026-05
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	period, err := time.Parse("2006-01", month)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "format month tidak valid")
		return
	}

	// Update overdue items
	_, _ = h.checklistRepo.UpdateOverdueItems(r.Context())

	// Ambil checklist
	items, err := h.checklistRepo.GenerateMonthlyChecklist(r.Context(), month)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil checklist")
		return
	}

	// Hitung total karyawan aktif
	// Catatan: salary belum ada di schema dasar, gunakan estimasi
	var totalEmployees int
	err = h.db.QueryRow(r.Context(), `
		SELECT COUNT(*)
		FROM users u
		WHERE u.is_active = true AND u.role != 'super_admin'
	`).Scan(&totalEmployees)
	if err != nil {
		totalEmployees = 0
	}
	// Estimasi total gaji bruto (placeholder — dalam implementasi nyata ambil dari payroll)
	totalGrossSalary := int64(totalEmployees) * 5_000_000

	// Kalkulasi BPJS total
	bpjsResult := CalculateBPJSForPeriod(totalGrossSalary, period)

	// Statistik checklist
	var pending, done, overdue int
	for _, item := range items {
		switch item.Status {
		case "done":
			done++
		case "overdue":
			overdue++
		default:
			pending++
		}
	}

	overallStatus := "green"
	if overdue > 0 {
		overallStatus = "red"
	} else {
		for _, item := range items {
			if item.Status == "pending" && item.DaysUntil <= 3 {
				overallStatus = "yellow"
				break
			}
		}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"period":         month,
		"overall_status": overallStatus,
		"checklist": map[string]interface{}{
			"items":   items,
			"pending": pending,
			"done":    done,
			"overdue": overdue,
		},
		"bpjs_summary": map[string]interface{}{
			"total_employees":           totalEmployees,
			"total_gross_salary":        totalGrossSalary,
			"total_company_contribution": bpjsResult.TotalCompanyContribution,
			"total_employee_deduction":  bpjsResult.TotalEmployeeContribution,
		},
		"upcoming_deadlines": getUpcomingDeadlines(items, 7),
	})
}

// SendDeadlineNotifications mengirim notifikasi H-3 untuk item yang mendekati deadline.
// Dipanggil oleh background worker.
func (h *Handler) SendDeadlineNotifications(ctx context.Context, hrAdminUserIDs []string) error {
	items, err := h.checklistRepo.GetItemsDueInDays(ctx, 3)
	if err != nil {
		return err
	}

	for _, item := range items {
		title := fmt.Sprintf("Deadline Compliance H-%d: %s", item.DaysUntil, item.Title)
		message := fmt.Sprintf(
			"Kewajiban '%s' akan jatuh tempo pada %s (%d hari lagi). Segera selesaikan.",
			item.Title,
			item.Deadline.Format("02 Jan 2006"),
			item.DaysUntil,
		)

		// Kirim ke semua HR Admin
		for _, userID := range hrAdminUserIDs {
			_ = h.notificationSvc.SendNotification(ctx, userID, "compliance_deadline", title, message, map[string]interface{}{
				"checklist_id": item.ID,
				"item_code":    item.ItemCode,
				"deadline":     item.Deadline,
				"days_until":   item.DaysUntil,
			})
		}

		// Tandai sudah dinotifikasi
		_ = h.checklistRepo.MarkNotifiedH3(ctx, item.ID)
	}

	return nil
}

// getUserSalary mengambil gaji karyawan dari database.
// Mengembalikan nilai default jika belum ada data salary.
func (h *Handler) getUserSalary(ctx context.Context, userID string) (int64, error) {
	// Placeholder: dalam implementasi nyata, ambil dari tabel payroll/profil karyawan
	// Untuk sekarang kembalikan nilai default
	_ = userID
	return 5_000_000, nil
}

// getUpcomingDeadlines mengambil item yang deadline-nya dalam N hari ke depan.
func getUpcomingDeadlines(items []*ChecklistItem, days int) []*ChecklistItem {
	var upcoming []*ChecklistItem
	for _, item := range items {
		if item.Status == "pending" && item.DaysUntil >= 0 && item.DaysUntil <= days {
			upcoming = append(upcoming, item)
		}
	}
	return upcoming
}
