package leave

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"saas-karyawan/internal/ai"
)

// NotificationSender adalah interface untuk mengirim notifikasi.
// Menggunakan interface agar tidak ada circular import dengan package notification.
type NotificationSender interface {
	SendNotification(ctx context.Context, userID, notifType, title, message string, metadata map[string]interface{}) error
}

// Service menangani logika bisnis untuk modul leave.
type Service struct {
	repo      *Repository
	aiClient  *ai.LLMClient
	notifSvc  NotificationSender
}

// NewService membuat instance Service baru.
func NewService(db *pgxpool.Pool, aiClient *ai.LLMClient) *Service {
	return &Service{
		repo:     NewRepository(db),
		aiClient: aiClient,
	}
}

// SetNotificationService menginjeksi notification service setelah inisialisasi.
// Dipanggil dari main.go setelah kedua service dibuat.
func (s *Service) SetNotificationService(ns NotificationSender) {
	s.notifSvc = ns
}

// GetLeaveTypes mengambil semua jenis cuti.
func (s *Service) GetLeaveTypes(ctx context.Context) ([]*LeaveType, error) {
	return s.repo.GetLeaveTypes(ctx)
}

// CreateLeaveRequest membuat pengajuan cuti baru.
func (s *Service) CreateLeaveRequest(ctx context.Context, userID string, req *CreateLeaveRequest) (*LeaveRequest, error) {
	if req.LeaveTypeID == "" {
		return nil, errors.New("jenis cuti wajib diisi")
	}
	if req.StartDate == "" || req.EndDate == "" {
		return nil, errors.New("tanggal mulai dan selesai wajib diisi")
	}
	if req.Reason == "" {
		return nil, errors.New("alasan cuti wajib diisi")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("format tanggal mulai tidak valid (gunakan YYYY-MM-DD)")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("format tanggal selesai tidak valid (gunakan YYYY-MM-DD)")
	}
	if endDate.Before(startDate) {
		return nil, errors.New("tanggal selesai tidak boleh sebelum tanggal mulai")
	}

	// Hitung total hari kerja (sederhana: selisih hari + 1)
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1

	leaveReq := &LeaveRequest{
		UserID:      userID,
		LeaveTypeID: req.LeaveTypeID,
		StartDate:   startDate,
		EndDate:     endDate,
		TotalDays:   totalDays,
		Reason:      req.Reason,
	}

	return s.repo.CreateLeaveRequest(ctx, leaveReq)
}

// GetLeaveRequests mengambil daftar pengajuan cuti.
// HR/Manager melihat semua, karyawan hanya milik sendiri.
func (s *Service) GetLeaveRequests(ctx context.Context, callerID, callerRole string, filter LeaveFilter) ([]*LeaveRequest, int, error) {
	// Karyawan biasa hanya bisa lihat milik sendiri
	if callerRole == "karyawan" {
		filter.UserID = callerID
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	return s.repo.List(ctx, filter)
}

// GetLeaveRequestByID mengambil detail satu pengajuan cuti.
func (s *Service) GetLeaveRequestByID(ctx context.Context, id, callerID, callerRole string) (*LeaveRequest, error) {
	req, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Karyawan hanya bisa lihat milik sendiri
	if callerRole == "karyawan" && req.UserID != callerID {
		return nil, errors.New("akses ditolak")
	}

	return req, nil
}

// ApproveLeaveRequest menyetujui pengajuan cuti.
func (s *Service) ApproveLeaveRequest(ctx context.Context, id, approverID string) (*LeaveRequest, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.Status != StatusPending {
		return nil, fmt.Errorf("pengajuan cuti tidak dalam status pending (status saat ini: %s)", existing.Status)
	}

	updated, err := s.repo.UpdateStatus(ctx, id, approverID, StatusApproved, "")
	if err != nil {
		return nil, err
	}

	// Kurangi saldo cuti
	year := existing.StartDate.Year()
	if err := s.repo.UpdateLeaveBalance(ctx, existing.UserID, existing.LeaveTypeID, year, existing.TotalDays); err != nil {
		// Log error tapi jangan gagalkan approve — balance bisa dikoreksi manual
		_ = err
	}

	// Kirim notifikasi ke pemohon
	if s.notifSvc != nil {
		_ = s.notifSvc.SendNotification(ctx, existing.UserID,
			"leave_approved",
			"Pengajuan cuti disetujui",
			"Pengajuan cuti Anda telah disetujui.",
			map[string]interface{}{"leave_id": id},
		)
	}

	return updated, nil
}

// RejectLeaveRequest menolak pengajuan cuti.
func (s *Service) RejectLeaveRequest(ctx context.Context, id, approverID, rejectionReason string) (*LeaveRequest, error) {
	if rejectionReason == "" {
		return nil, errors.New("alasan penolakan wajib diisi")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.Status != StatusPending {
		return nil, fmt.Errorf("pengajuan cuti tidak dalam status pending (status saat ini: %s)", existing.Status)
	}

	updated, err := s.repo.UpdateStatus(ctx, id, approverID, StatusRejected, rejectionReason)
	if err != nil {
		return nil, err
	}

	// Kirim notifikasi ke pemohon
	if s.notifSvc != nil {
		_ = s.notifSvc.SendNotification(ctx, existing.UserID,
			"leave_rejected",
			"Pengajuan cuti ditolak",
			fmt.Sprintf("Pengajuan cuti Anda ditolak. Alasan: %s", rejectionReason),
			map[string]interface{}{"leave_id": id},
		)
	}

	return updated, nil
}

// CancelLeaveRequest membatalkan pengajuan cuti (hanya owner, hanya jika pending).
func (s *Service) CancelLeaveRequest(ctx context.Context, id, callerID string) (*LeaveRequest, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.UserID != callerID {
		return nil, errors.New("hanya pemilik pengajuan yang bisa membatalkan")
	}
	if existing.Status != StatusPending {
		return nil, fmt.Errorf("hanya pengajuan dengan status pending yang bisa dibatalkan (status saat ini: %s)", existing.Status)
	}

	return s.repo.UpdateStatus(ctx, id, "", StatusCancelled, "")
}

// GetLeaveBalance mengambil saldo cuti user untuk tahun tertentu.
func (s *Service) GetLeaveBalance(ctx context.Context, userID string, year int) ([]*LeaveBalance, error) {
	if year <= 0 {
		year = time.Now().Year()
	}
	return s.repo.GetLeaveBalance(ctx, userID, year)
}

// GetAILeaveRecommendation menghasilkan rekomendasi AI untuk pengajuan cuti.
func (s *Service) GetAILeaveRecommendation(ctx context.Context, requestID string) (*AIRecommendation, error) {
	if s.aiClient == nil {
		return &AIRecommendation{
			Recommendation: "Perlu pertimbangan",
			Reason:         "AI tidak tersedia saat ini",
		}, nil
	}

	req, err := s.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Ambil histori cuti tim 3 bulan terakhir
	history, err := s.repo.GetTeamLeaveHistory(ctx, 3)
	if err != nil {
		history = []map[string]interface{}{}
	}

	prompt := s.buildAIPrompt(req, history)

	messages := []ai.ChatMessage{
		{
			Role:    "system",
			Content: "Kamu adalah asisten HR yang membantu mengevaluasi pengajuan cuti karyawan. Berikan rekomendasi singkat dalam Bahasa Indonesia. Jawab hanya dengan format JSON: {\"recommendation\": \"Direkomendasikan disetujui\" atau \"Perlu pertimbangan\", \"reason\": \"alasan singkat max 2 kalimat\"}",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := s.aiClient.Chat(ctx, messages)
	if err != nil {
		return &AIRecommendation{
			Recommendation: "Perlu pertimbangan",
			Reason:         "Tidak dapat menghubungi AI saat ini",
		}, nil
	}

	// Parse JSON response dari AI
	rec := parseAIRecommendation(response)
	return rec, nil
}

func (s *Service) buildAIPrompt(req *LeaveRequest, history []map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Evaluasi pengajuan cuti berikut:\n"))
	sb.WriteString(fmt.Sprintf("- Karyawan: %s\n", req.UserName))
	sb.WriteString(fmt.Sprintf("- Jenis Cuti: %s\n", req.LeaveTypeName))
	sb.WriteString(fmt.Sprintf("- Tanggal: %s s/d %s (%d hari)\n",
		req.StartDate.Format("2006-01-02"),
		req.EndDate.Format("2006-01-02"),
		req.TotalDays,
	))
	sb.WriteString(fmt.Sprintf("- Alasan: %s\n\n", req.Reason))

	if len(history) > 0 {
		sb.WriteString("Histori cuti tim 3 bulan terakhir:\n")
		for _, h := range history {
			sb.WriteString(fmt.Sprintf("- %s: %s (%s s/d %s, %v hari) - %s\n",
				h["name"], h["leave_type"], h["start_date"], h["end_date"], h["total_days"], h["status"],
			))
		}
	}

	sb.WriteString("\nBerikan rekomendasi apakah pengajuan ini layak disetujui berdasarkan konteks tim.")
	return sb.String()
}

func parseAIRecommendation(response string) *AIRecommendation {
	// Cari JSON dalam response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return &AIRecommendation{
			Recommendation: "Perlu pertimbangan",
			Reason:         response,
		}
	}

	jsonStr := response[start : end+1]

	// Parse manual untuk menghindari import encoding/json yang circular
	rec := &AIRecommendation{}
	if strings.Contains(jsonStr, "Direkomendasikan disetujui") {
		rec.Recommendation = "Direkomendasikan disetujui"
	} else {
		rec.Recommendation = "Perlu pertimbangan"
	}

	// Ambil reason
	reasonStart := strings.Index(jsonStr, `"reason"`)
	if reasonStart != -1 {
		afterKey := jsonStr[reasonStart+len(`"reason"`):]
		colonIdx := strings.Index(afterKey, ":")
		if colonIdx != -1 {
			afterColon := strings.TrimSpace(afterKey[colonIdx+1:])
			if len(afterColon) > 0 && afterColon[0] == '"' {
				endQuote := strings.Index(afterColon[1:], "\"")
				if endQuote != -1 {
					rec.Reason = afterColon[1 : endQuote+1]
				}
			}
		}
	}

	if rec.Reason == "" {
		rec.Reason = "Berdasarkan analisis data tim"
	}

	return rec
}
