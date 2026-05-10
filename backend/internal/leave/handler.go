package leave

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"saas-karyawan/internal/auth"
	"saas-karyawan/pkg/response"
)

// Handler menangani HTTP request untuk modul leave.
type Handler struct {
	svc *Service
}

// NewHandler membuat instance Handler baru.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetLeaveTypes menangani GET /api/leaves/types
func (h *Handler) GetLeaveTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.svc.GetLeaveTypes(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil jenis cuti")
		return
	}
	response.Success(w, "jenis cuti berhasil diambil", types)
}

// Create menangani POST /api/leaves
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	var req CreateLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "request body tidak valid")
		return
	}

	leaveReq, err := h.svc.CreateLeaveRequest(r.Context(), claims.UserID, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(w, "pengajuan cuti berhasil dibuat", leaveReq)
}

// List menangani GET /api/leaves
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	filter := LeaveFilter{
		UserID:    r.URL.Query().Get("user_id"),
		Status:    r.URL.Query().Get("status"),
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
		Page:      page,
		PageSize:  pageSize,
	}

	requests, total, err := h.svc.GetLeaveRequests(r.Context(), claims.UserID, claims.Role, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil daftar cuti")
		return
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	response.Paginated(w, "daftar cuti berhasil diambil", requests, total, filter.Page, filter.PageSize)
}

// GetByID menangani GET /api/leaves/:id
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetClaims(r)

	req, err := h.svc.GetLeaveRequestByID(r.Context(), id, claims.UserID, claims.Role)
	if err != nil {
		if err.Error() == "akses ditolak" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusNotFound, "pengajuan cuti tidak ditemukan")
		return
	}

	response.Success(w, "detail cuti berhasil diambil", req)
}

// Approve menangani PUT /api/leaves/:id/approve
func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetClaims(r)

	req, err := h.svc.ApproveLeaveRequest(r.Context(), id, claims.UserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, "pengajuan cuti berhasil disetujui", req)
}

// Reject menangani PUT /api/leaves/:id/reject
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetClaims(r)

	var req RejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "request body tidak valid")
		return
	}

	leaveReq, err := h.svc.RejectLeaveRequest(r.Context(), id, claims.UserID, req.RejectionReason)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, "pengajuan cuti berhasil ditolak", leaveReq)
}

// Cancel menangani PUT /api/leaves/:id/cancel
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetClaims(r)

	req, err := h.svc.CancelLeaveRequest(r.Context(), id, claims.UserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, "pengajuan cuti berhasil dibatalkan", req)
}

// GetMyBalance menangani GET /api/leaves/balance
func (h *Handler) GetMyBalance(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)

	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	balances, err := h.svc.GetLeaveBalance(r.Context(), claims.UserID, year)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil saldo cuti")
		return
	}

	response.Success(w, "saldo cuti berhasil diambil", balances)
}

// GetBalanceByUserID menangani GET /api/leaves/balance/:user_id (HR only)
func (h *Handler) GetBalanceByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	balances, err := h.svc.GetLeaveBalance(r.Context(), userID, year)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mengambil saldo cuti")
		return
	}

	response.Success(w, "saldo cuti karyawan berhasil diambil", balances)
}

// GetAIRecommendation menangani GET /api/leaves/:id/ai-recommendation
func (h *Handler) GetAIRecommendation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rec, err := h.svc.GetAILeaveRecommendation(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "gagal mendapatkan rekomendasi AI")
		return
	}

	response.Success(w, "rekomendasi AI berhasil diambil", rec)
}
