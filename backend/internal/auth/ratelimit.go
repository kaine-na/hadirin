package auth

import (
	"net/http"
	"sync"
	"time"

	"saas-karyawan/pkg/response"
)

// loginAttempt menyimpan data percobaan login per IP.
type loginAttempt struct {
	count     int
	windowEnd time.Time
}

// LoginRateLimiter membatasi percobaan login per IP: max 5 request/menit.
type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
	max      int
	window   time.Duration
}

// NewLoginRateLimiter membuat rate limiter baru.
func NewLoginRateLimiter() *LoginRateLimiter {
	rl := &LoginRateLimiter{
		attempts: make(map[string]*loginAttempt),
		max:      5,
		window:   time.Minute,
	}
	// Bersihkan entri lama setiap 5 menit agar tidak memory leak.
	go rl.cleanup()
	return rl
}

// Allow memeriksa apakah IP diizinkan melakukan request.
// Mengembalikan true jika masih dalam batas, false jika sudah melebihi.
func (rl *LoginRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	a, ok := rl.attempts[ip]
	if !ok || now.After(a.windowEnd) {
		// Window baru atau IP baru
		rl.attempts[ip] = &loginAttempt{
			count:     1,
			windowEnd: now.Add(rl.window),
		}
		return true
	}

	if a.count >= rl.max {
		return false
	}

	a.count++
	return true
}

// cleanup menghapus entri yang sudah expired secara periodik.
func (rl *LoginRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, a := range rl.attempts {
			if now.After(a.windowEnd) {
				delete(rl.attempts, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware mengembalikan HTTP middleware yang menerapkan rate limiting.
func (rl *LoginRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		if !rl.Allow(ip) {
			response.Error(w, http.StatusTooManyRequests,
				"terlalu banyak percobaan login, coba lagi dalam 1 menit")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// realIP mengekstrak IP asli dari request, mempertimbangkan proxy header.
func realIP(r *http.Request) string {
	// Cek X-Forwarded-For (dari reverse proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Ambil IP pertama (client asli)
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	// Cek X-Real-IP (dari nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fallback ke RemoteAddr, strip port
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
