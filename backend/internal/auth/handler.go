package auth

import (
	"encoding/json"
	"net/http"

	"saas-karyawan/pkg/response"
)

// Handler menangani HTTP request untuk modul auth.
type Handler struct {
	svc *Service
}

// NewHandler membuat instance Handler baru.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Login menangani POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "request body tidak valid")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "email dan password wajib diisi")
		return
	}

	tokenResp, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(w, "login berhasil", tokenResp)
}

// Logout menangani POST /api/auth/logout
// Token invalidation dilakukan di sisi client (hapus token dari storage).
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "logout berhasil", nil)
}

// Me menangani GET /api/auth/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	user, err := h.svc.GetMe(r.Context(), claims.UserID)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.Success(w, "data user berhasil diambil", user)
}
