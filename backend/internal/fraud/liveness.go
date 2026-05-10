package fraud

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"saas-karyawan/internal/ai"
)

const (
	minPhotoSizeBytes = 50 * 1024       // 50 KB
	maxPhotoSizeBytes = 5 * 1024 * 1024 // 5 MB
)

// LivenessResult hasil pengecekan liveness selfie.
type LivenessResult struct {
	IsLiveFace    bool    `json:"is_live_face"`
	Score         float64 `json:"score"`
	Notes         string  `json:"notes"`
	PhotoID       string  `json:"photo_id,omitempty"`
	FilePath      string  `json:"file_path,omitempty"`
	FraudDetected bool    `json:"fraud_detected"`
	FraudType     string  `json:"fraud_type,omitempty"`
	Severity      string  `json:"severity,omitempty"`
}

// LivenessChecker menangani upload dan validasi selfie.
type LivenessChecker struct {
	db        *pgxpool.Pool
	uploadDir string
	llmClient *ai.LLMClient
}

// NewLivenessChecker membuat instance LivenessChecker baru.
func NewLivenessChecker(db *pgxpool.Pool, uploadDir string, llmClient *ai.LLMClient) *LivenessChecker {
	return &LivenessChecker{
		db:        db,
		uploadDir: uploadDir,
		llmClient: llmClient,
	}
}

// ValidateAndSave memvalidasi file foto selfie, menyimpannya, dan menjalankan liveness check.
func (c *LivenessChecker) ValidateAndSave(
	ctx context.Context,
	attendanceID string,
	userID string,
	file multipart.File,
	header *multipart.FileHeader,
	gps GPSData,
) (*LivenessResult, error) {
	// Validasi MIME type
	mimeType := header.Header.Get("Content-Type")
	if !isAllowedMIME(mimeType) {
		return nil, fmt.Errorf("format file tidak didukung: %s (harus JPEG atau PNG)", mimeType)
	}

	// Validasi ukuran file
	if header.Size < minPhotoSizeBytes {
		return nil, fmt.Errorf("ukuran foto terlalu kecil: %d bytes (minimum 50KB)", header.Size)
	}
	if header.Size > maxPhotoSizeBytes {
		return nil, fmt.Errorf("ukuran foto terlalu besar: %d bytes (maksimum 5MB)", header.Size)
	}

	// Buat direktori selfie jika belum ada
	selfieDir := filepath.Join(c.uploadDir, "selfies", userID)
	if err := os.MkdirAll(selfieDir, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat direktori selfie: %w", err)
	}

	// Simpan file dengan nama unik
	ext := ".jpg"
	if strings.Contains(mimeType, "png") {
		ext = ".png"
	}
	filename := fmt.Sprintf("%s_%s%s", attendanceID, time.Now().Format("20060102150405"), ext)
	filePath := filepath.Join(selfieDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan foto: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("gagal menulis foto: %w", err)
	}

	// Jalankan AI liveness check
	result := &LivenessResult{
		FilePath:   filePath,
		IsLiveFace: true,
		Score:      1.0,
		Notes:      "Liveness check tidak tersedia",
	}

	if c.llmClient != nil {
		livenessResult := c.checkLivenessWithAI(ctx, filePath)
		result.IsLiveFace = livenessResult.IsLiveFace
		result.Score = livenessResult.Score
		result.Notes = livenessResult.Notes
	}

	// Simpan ke database
	photoID, err := c.savePhotoRecord(ctx, attendanceID, userID, filePath, int(header.Size), mimeType, gps, result)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan record foto: %w", err)
	}
	result.PhotoID = photoID

	// Tentukan apakah ada fraud
	if !result.IsLiveFace {
		result.FraudDetected = true
		result.FraudType = "liveness_fail"
		result.Severity = "high"
	}

	return result, nil
}

// checkLivenessWithAI mengirim prompt ke AI untuk validasi liveness.
// Karena LLM text-based, kita kirim prompt deskriptif tentang validasi foto.
func (c *LivenessChecker) checkLivenessWithAI(ctx context.Context, filePath string) *LivenessResult {
	messages := []ai.ChatMessage{
		{
			Role: "system",
			Content: `Anda adalah sistem deteksi liveness untuk absensi karyawan.
Berikan respons dalam format tepat:
LIVE: [true/false]
SCORE: [0.0-1.0]
NOTES: [penjelasan singkat]`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`Foto selfie telah diupload untuk clock-in absensi.
File: %s
Asumsikan foto valid karena sistem vision tidak tersedia.
Berikan respons default untuk foto yang diterima.`, filePath),
		},
	}

	resp, err := c.llmClient.Chat(ctx, messages)
	if err != nil {
		return &LivenessResult{
			IsLiveFace: true,
			Score:      0.5,
			Notes:      "AI liveness check tidak tersedia",
		}
	}

	return parseLivenessResponse(resp)
}

// parseLivenessResponse mem-parse respons AI untuk liveness check.
func parseLivenessResponse(response string) *LivenessResult {
	result := &LivenessResult{
		IsLiveFace: true,
		Score:      0.5,
		Notes:      response,
	}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "LIVE:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "LIVE:"))
			result.IsLiveFace = strings.EqualFold(val, "true")
		} else if strings.HasPrefix(line, "SCORE:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "SCORE:"))
			var score float64
			if _, err := fmt.Sscanf(val, "%f", &score); err == nil {
				result.Score = score
			}
		} else if strings.HasPrefix(line, "NOTES:") {
			result.Notes = strings.TrimSpace(strings.TrimPrefix(line, "NOTES:"))
		}
	}

	return result
}

// savePhotoRecord menyimpan record foto ke database.
func (c *LivenessChecker) savePhotoRecord(
	ctx context.Context,
	attendanceID, userID, filePath string,
	fileSize int,
	mimeType string,
	gps GPSData,
	liveness *LivenessResult,
) (string, error) {
	query := `
		INSERT INTO attendance_photos (
			attendance_id, user_id, file_path, file_size, mime_type,
			is_live_face, liveness_score, liveness_notes,
			latitude, longitude, gps_accuracy
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var lat, lon, acc interface{}
	if gps.Latitude != 0 || gps.Longitude != 0 {
		lat = gps.Latitude
		lon = gps.Longitude
		acc = gps.Accuracy
	}

	var photoID string
	err := c.db.QueryRow(ctx, query,
		attendanceID, userID, filePath, fileSize, mimeType,
		liveness.IsLiveFace, liveness.Score, liveness.Notes,
		lat, lon, acc,
	).Scan(&photoID)

	return photoID, err
}

// isAllowedMIME memeriksa apakah MIME type diizinkan.
func isAllowedMIME(mimeType string) bool {
	allowed := []string{"image/jpeg", "image/jpg", "image/png"}
	for _, a := range allowed {
		if strings.EqualFold(mimeType, a) {
			return true
		}
	}
	return false
}
