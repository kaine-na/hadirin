package fraud

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"saas-karyawan/internal/ai"
	"saas-karyawan/internal/auth"
	"saas-karyawan/pkg/response"
)

// Handler menangani HTTP request untuk modul fraud detection.
type Handler struct {
	repo      *Repository
	gpsVal    *GPSValidator
	liveness  *LivenessChecker
	anomaly   *AnomalyDetector
	uploadDir string
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo *Repository, gpsVal *GPSValidator, liveness *LivenessChecker, anomaly *AnomalyDetector, uploadDir string) *Handler {
	return &Handler{
		repo:      repo,
		gpsVal:    gpsVal,
		liveness:  liveness,
		anomaly:   anomaly,
		uploadDir: uploadDir,
	}
}

// NewHandlerFromDeps membuat Handler dari dependencies yang sudah ada.
func NewHandlerFromDeps(db interface{ QueryRow(ctx interface{}, sql string, args ...interface{}) interface{ Scan(dest ...interface{}) error } }, llmClient *ai.LLMClient, uploadDir string) *Handler {
	return nil // placeholder — gunakan NewHandler langsung
}

// ClockInWithFraudCheck menangani POST /api/attendance/clock-in dengan fraud detection.
// Ini adalah handler tambahan yang dipanggil dari attendance handler.
func (h *Handler) ValidateClockIn(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	// Parse multipart form (foto selfie + data GPS)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "gagal parse form data")
		return
	}

	// Ambil data GPS dari form
	gps := GPSData{
		UserAgent: r.UserAgent(),
	}
	if lat := r.FormValue("latitude"); lat != "" {
		fmt.Sscanf(lat, "%f", &gps.Latitude)
	}
	if lon := r.FormValue("longitude"); lon != "" {
		fmt.Sscanf(lon, "%f", &gps.Longitude)
	}
	if acc := r.FormValue("accuracy"); acc != "" {
		fmt.Sscanf(acc, "%f", &gps.Accuracy)
	}

	// Ambil attendance_id dari form (sudah dibuat oleh attendance handler)
	attendanceID := r.FormValue("attendance_id")

	var fraudResults []map[string]interface{}

	// 1. Validasi GPS accuracy
	if accResult := h.gpsVal.ValidateAccuracy(gps); !accResult.Valid {
		fraudResults = append(fraudResults, map[string]interface{}{
			"type":        accResult.FraudType,
			"severity":    accResult.Severity,
			"description": accResult.Description,
			"evidence":    accResult.Evidence,
		})
		h.saveFraudLog(r, claims.UserID, attendanceID, accResult)
	}

	// 2. Deteksi mock location
	if mockResult := h.gpsVal.DetectMockLocation(gps); !mockResult.Valid {
		fraudResults = append(fraudResults, map[string]interface{}{
			"type":        mockResult.FraudType,
			"severity":    mockResult.Severity,
			"description": mockResult.Description,
			"evidence":    mockResult.Evidence,
		})
		h.saveFraudLog(r, claims.UserID, attendanceID, mockResult)
	}

	// 3. Velocity check
	if velResult, err := h.gpsVal.CheckVelocity(r.Context(), claims.UserID, gps); err == nil && !velResult.Valid {
		fraudResults = append(fraudResults, map[string]interface{}{
			"type":        velResult.FraudType,
			"severity":    velResult.Severity,
			"description": velResult.Description,
			"evidence":    velResult.Evidence,
		})
		h.saveFraudLog(r, claims.UserID, attendanceID, velResult)
	}

	// 4. Anomaly detection
	deviceHash := generateDeviceHash(r.UserAgent(), r.Header.Get("Accept-Language"))
	if deviceHash != "" {
		_ = h.repo.UpsertDeviceFingerprint(r.Context(), claims.UserID, deviceHash, r.UserAgent())
	}

	anomalies, _ := h.anomaly.Analyze(r.Context(), claims.UserID, time.Now(), gps, deviceHash)
	for _, a := range anomalies {
		fraudResults = append(fraudResults, map[string]interface{}{
			"type":        a.FraudType,
			"severity":    a.Severity,
			"description": a.Description,
			"evidence":    a.Evidence,
			"ai_analysis": a.AIAnalysis,
		})
		h.saveFraudLogFromAnomaly(r, claims.UserID, attendanceID, a)
	}

	// 5. Selfie liveness check (jika ada foto)
	var livenessResult *LivenessResult
	file, header, err := r.FormFile("selfie")
	if err == nil {
		defer file.Close()
		if attendanceID != "" {
			livenessResult, err = h.liveness.ValidateAndSave(r.Context(), attendanceID, claims.UserID, file, header, gps)
			if err == nil && livenessResult.FraudDetected {
				fraudResults = append(fraudResults, map[string]interface{}{
					"type":        livenessResult.FraudType,
					"severity":    livenessResult.Severity,
					"description": "Selfie liveness check gagal",
					"evidence": Evidence{
						"liveness_score": livenessResult.Score,
						"notes":          livenessResult.Notes,
					},
				})
			}
		}
	}

	result := map[string]interface{}{
		"fraud_detected": len(fraudResults) > 0,
		"fraud_count":    len(fraudResults),
		"fraud_results":  fraudResults,
	}

	if livenessResult != nil {
		result["liveness"] = livenessResult
	}

	response.Success(w, "validasi fraud selesai", result)
}

// ListFraudLogs menangani GET /api/fraud/logs (HR only)
func (h *Handler) ListFraudLogs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	logs, total, err := h.repo.ListFraudLogs(r.Context(), status, page, pageSize)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil fraud logs")
		return
	}

	response.Paginated(w, "fraud logs berhasil diambil", logs, total, page, pageSize)
}

// GetFraudLogByID menangani GET /api/fraud/logs/:id (HR only)
func (h *Handler) GetFraudLogByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	log, err := h.repo.GetFraudLogByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "fraud log tidak ditemukan")
		return
	}

	response.Success(w, "detail fraud log", log)
}

// DismissFraudLog menangani PUT /api/fraud/logs/:id/dismiss (HR only)
func (h *Handler) DismissFraudLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetClaims(r)

	var req ReviewRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.repo.UpdateFraudLogStatus(r.Context(), id, "dismissed", claims.UserID, req.Notes); err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal dismiss fraud log")
		return
	}

	response.Success(w, "fraud log berhasil di-dismiss", nil)
}

// ConfirmFraudLog menangani PUT /api/fraud/logs/:id/confirm (HR only)
func (h *Handler) ConfirmFraudLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetClaims(r)

	var req ReviewRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.repo.UpdateFraudLogStatus(r.Context(), id, "confirmed", claims.UserID, req.Notes); err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal konfirmasi fraud log")
		return
	}

	response.Success(w, "fraud log berhasil dikonfirmasi", nil)
}

// GetFraudSummary menangani GET /api/fraud/summary (HR only)
func (h *Handler) GetFraudSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.repo.GetFraudSummary(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil ringkasan fraud")
		return
	}

	response.Success(w, "ringkasan fraud bulan ini", summary)
}

// saveFraudLog menyimpan fraud log dari hasil GPS validation.
func (h *Handler) saveFraudLog(r *http.Request, userID, attendanceID string, result *GPSValidationResult) {
	log := &FraudLog{
		UserID:      userID,
		FraudType:   result.FraudType,
		Severity:    result.Severity,
		Description: result.Description,
		Evidence:    result.Evidence,
	}
	if attendanceID != "" {
		log.AttendanceID = &attendanceID
	}
	_, _ = h.repo.CreateFraudLog(r.Context(), log)
}

// saveFraudLogFromAnomaly menyimpan fraud log dari hasil anomaly detection.
func (h *Handler) saveFraudLogFromAnomaly(r *http.Request, userID, attendanceID string, anomaly AnomalyResult) {
	log := &FraudLog{
		UserID:      userID,
		FraudType:   anomaly.FraudType,
		Severity:    anomaly.Severity,
		Description: anomaly.Description,
		Evidence:    anomaly.Evidence,
		AIAnalysis:  anomaly.AIAnalysis,
	}
	if attendanceID != "" {
		log.AttendanceID = &attendanceID
	}
	_, _ = h.repo.CreateFraudLog(r.Context(), log)
}

// generateDeviceHash membuat hash sederhana dari user-agent dan accept-language.
func generateDeviceHash(userAgent, acceptLang string) string {
	if userAgent == "" {
		return ""
	}
	data := userAgent + "|" + acceptLang
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}
