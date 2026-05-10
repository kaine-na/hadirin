package notification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"saas-karyawan/internal/auth"
)

// Handler menangani HTTP request untuk modul notification.
type Handler struct {
	svc *Service
}

// NewHandler membuat instance Handler baru.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Stream menangani SSE endpoint untuk real-time push notifikasi.
// GET /api/notifications/stream
func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming tidak didukung", http.StatusInternalServerError)
		return
	}

	// Daftarkan client ke hub
	ch := h.svc.hub.Register(userID)
	defer h.svc.hub.Unregister(userID, ch)

	// Kirim event "connected" sebagai handshake awal
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	// Loop: tunggu event atau client disconnect
	for {
		select {
		case <-r.Context().Done():
			// Client disconnect
			return
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// List mengambil daftar notifikasi (paginated).
// GET /api/notifications
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	userID := claims.UserID

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	filter := ListFilter{
		Page:       page,
		PageSize:   pageSize,
		UnreadOnly: unreadOnly,
	}

	notifications, total, err := h.svc.List(r.Context(), userID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"notifications": notifications,
			"total":         total,
			"page":          filter.Page,
			"page_size":     filter.PageSize,
		},
	})
}

// MarkAsRead menandai satu notifikasi sebagai sudah dibaca.
// PUT /api/notifications/:id/read
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	userID := claims.UserID
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id notifikasi wajib diisi")
		return
	}

	if err := h.svc.MarkAsRead(r.Context(), id, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Notifikasi ditandai sudah dibaca",
	})
}

// MarkAllAsRead menandai semua notifikasi user sebagai sudah dibaca.
// PUT /api/notifications/read-all
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	userID := claims.UserID

	if err := h.svc.MarkAllAsRead(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Semua notifikasi ditandai sudah dibaca",
	})
}

// GetUnreadCount mengambil jumlah notifikasi yang belum dibaca.
// GET /api/notifications/unread-count
func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	userID := claims.UserID

	count, err := h.svc.GetUnreadCount(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    UnreadCountResponse{Count: count},
	})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]interface{}{
		"success": false,
		"message": msg,
	})
}
