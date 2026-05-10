package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// contextKey adalah tipe untuk key di context agar tidak konflik.
type contextKey string

const (
	// ClaimsKey adalah key untuk menyimpan JWT claims di request context.
	ClaimsKey contextKey = "claims"
)

// Service menangani logika bisnis autentikasi.
type Service struct {
	db          *pgxpool.Pool
	jwtSecret   []byte
	expiryHours int
}

// NewService membuat instance Service baru.
func NewService(db *pgxpool.Pool, jwtSecret string, expiryHours int) *Service {
	return &Service{
		db:          db,
		jwtSecret:   []byte(jwtSecret),
		expiryHours: expiryHours,
	}
}

// Login memvalidasi kredensial dan mengembalikan JWT token.
func (s *Service) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	var (
		id           string
		name         string
		userEmail    string
		passwordHash string
		role         string
		department   string
		position     string
		photoURL     string
		companyID    string
		isActive     bool
	)

	err := s.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, 
		       COALESCE(department, ''), COALESCE(position, ''), 
		       COALESCE(photo_url, ''), company_id::text, is_active
		FROM users 
		WHERE email = $1
	`, email).Scan(
		&id, &name, &userEmail, &passwordHash,
		&role, &department, &position, &photoURL,
		&companyID, &isActive,
	)
	if err != nil {
		// Jangan bocorkan apakah email ada atau tidak
		return nil, errors.New("email atau password salah")
	}

	if !isActive {
		return nil, errors.New("akun tidak aktif")
	}

	// Verifikasi password dengan bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	// Generate JWT token
	expiresAt := time.Now().Add(time.Duration(s.expiryHours) * time.Hour)
	token, err := s.generateToken(id, userEmail, role, companyID, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:         id,
			Name:       name,
			Email:      userEmail,
			Role:       role,
			Department: department,
			Position:   position,
			PhotoURL:   photoURL,
		},
	}, nil
}

// GetMe mengambil data user berdasarkan ID dari claims.
func (s *Service) GetMe(ctx context.Context, userID string) (*UserInfo, error) {
	var (
		id         string
		name       string
		email      string
		role       string
		department string
		position   string
		photoURL   string
	)

	err := s.db.QueryRow(ctx, `
		SELECT id, name, email, role,
		       COALESCE(department, ''), COALESCE(position, ''), COALESCE(photo_url, '')
		FROM users 
		WHERE id = $1 AND is_active = true
	`, userID).Scan(&id, &name, &email, &role, &department, &position, &photoURL)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	return &UserInfo{
		ID:         id,
		Name:       name,
		Email:      email,
		Role:       role,
		Department: department,
		Position:   position,
		PhotoURL:   photoURL,
	}, nil
}

// generateToken membuat JWT token dengan claims yang diberikan.
func (s *Service) generateToken(userID, email, role, companyID string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"email":      email,
		"role":       role,
		"company_id": companyID,
		"exp":        expiresAt.Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken memvalidasi JWT token dan mengembalikan claims.
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Gunakan comma-ok pattern untuk menghindari panic jika klaim tidak ada atau bukan string.
	userID, ok1 := mapClaims["user_id"].(string)
	email, ok2 := mapClaims["email"].(string)
	role, ok3 := mapClaims["role"].(string)
	companyID, ok4 := mapClaims["company_id"].(string)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return nil, errors.New("token claims tidak lengkap atau tipe tidak valid")
	}

	return &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		CompanyID: companyID,
	}, nil
}
