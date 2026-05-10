package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config menyimpan semua konfigurasi aplikasi dari environment variables.
type Config struct {
	DatabaseURL        string
	JWTSecret          string
	JWTExpiryHours     int
	UploadDir          string
	MaxFileSizeMB      int64
	LLMBaseURL         string
	LLMAPIKey          string
	LLMModel           string
	LLMTimeoutSeconds  int
	Port               string
	CORSOrigins        string
}

// Load membaca .env file dan mengembalikan Config.
// Jika .env tidak ada, tetap lanjut dengan env vars yang sudah ada.
func Load() *Config {
	// Abaikan error jika .env tidak ada (production pakai env vars langsung)
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/saas_karyawan?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTExpiryHours:    getEnvInt("JWT_EXPIRY_HOURS", 24),
		UploadDir:         getEnv("UPLOAD_DIR", "./uploads"),
		MaxFileSizeMB:     int64(getEnvInt("MAX_FILE_SIZE_MB", 10)),
		LLMBaseURL:        getEnv("LLM_BASE_URL", "http://43.133.61.163:8787/v1"),
		LLMAPIKey:         getEnv("LLM_API_KEY", "sk-pool"),
		LLMModel:          getEnv("LLM_MODEL", "claude-sonnet-4.6"),
		LLMTimeoutSeconds: getEnvInt("LLM_TIMEOUT_SECONDS", 60),
		Port:              getEnv("PORT", "8080"),
		CORSOrigins:       getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
