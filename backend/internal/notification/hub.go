package notification

import (
	"encoding/json"
	"log"
	"sync"
)

// sseClient merepresentasikan satu koneksi SSE dari browser.
type sseClient struct {
	userID string
	ch     chan []byte
}

// SSEHub mengelola semua koneksi SSE aktif.
// Setiap user bisa punya lebih dari satu tab/koneksi.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[string][]*sseClient // userID -> list of clients
}

// NewSSEHub membuat SSEHub baru.
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[string][]*sseClient),
	}
}

// Register mendaftarkan client baru dan mengembalikan channel untuk menerima event.
func (h *SSEHub) Register(userID string) chan []byte {
	ch := make(chan []byte, 16) // buffer 16 event
	client := &sseClient{userID: userID, ch: ch}

	h.mu.Lock()
	h.clients[userID] = append(h.clients[userID], client)
	h.mu.Unlock()

	return ch
}

// Unregister menghapus client dari hub saat koneksi SSE ditutup.
func (h *SSEHub) Unregister(userID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.clients[userID]
	for i, c := range clients {
		if c.ch == ch {
			// Hapus dari slice tanpa mengubah urutan
			h.clients[userID] = append(clients[:i], clients[i+1:]...)
			close(ch)
			break
		}
	}

	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
}

// Broadcast mengirim notifikasi ke semua koneksi SSE milik user tertentu.
// Non-blocking: jika channel penuh, event di-drop (tidak memblokir goroutine).
func (h *SSEHub) Broadcast(userID string, n *Notification) {
	data, err := json.Marshal(n)
	if err != nil {
		log.Printf("[SSEHub] gagal marshal notifikasi: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	for _, c := range clients {
		select {
		case c.ch <- data:
		default:
			// Channel penuh, skip — jangan blokir
			log.Printf("[SSEHub] channel penuh untuk user %s, event di-drop", userID)
		}
	}
}
