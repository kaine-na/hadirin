package auth

import (
	"context"
	"net/http"
	"strings"

	"saas-karyawan/pkg/response"
)

// RequireAuth adalah middleware yang memvalidasi JWT token dari header Authorization.
// Claims diinjeksi ke request context dengan key ClaimsKey.
func RequireAuth(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "token tidak ditemukan")
				return
			}

			// Format: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				response.Error(w, http.StatusUnauthorized, "format token tidak valid")
				return
			}

			claims, err := svc.ValidateToken(parts[1])
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "token tidak valid atau sudah expired")
				return
			}

			// Inject claims ke context
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole adalah middleware yang memastikan user memiliki salah satu role yang diizinkan.
// Harus digunakan setelah RequireAuth.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowedRoles := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowedRoles[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsKey).(*Claims)
			if !ok || claims == nil {
				response.Error(w, http.StatusUnauthorized, "tidak terautentikasi")
				return
			}

			if !allowedRoles[claims.Role] {
				response.Error(w, http.StatusForbidden, "akses ditolak: role tidak memiliki izin")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClaims mengambil Claims dari request context.
// Mengembalikan nil jika tidak ada (endpoint tidak dilindungi middleware).
func GetClaims(r *http.Request) *Claims {
	claims, _ := r.Context().Value(ClaimsKey).(*Claims)
	return claims
}
