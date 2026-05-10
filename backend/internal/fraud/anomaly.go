package fraud

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"saas-karyawan/internal/ai"
)

// AnomalyResult hasil deteksi anomali pola absensi.
type AnomalyResult struct {
	HasAnomaly  bool     `json:"has_anomaly"`
	FraudType   string   `json:"fraud_type,omitempty"`
	Severity    string   `json:"severity,omitempty"`
	Description string   `json:"description,omitempty"`
	Evidence    Evidence `json:"evidence,omitempty"`
	AIAnalysis  string   `json:"ai_analysis,omitempty"`
}

// AttendancePattern pola absensi karyawan dari 7 hari terakhir.
type AttendancePattern struct {
	AvgClockInHour   float64
	MedianLatitude   float64
	MedianLongitude  float64
	CommonDeviceHash string
	SampleCount      int
}

// AnomalyDetector mendeteksi anomali pola absensi.
type AnomalyDetector struct {
	db        *pgxpool.Pool
	llmClient *ai.LLMClient
}

// NewAnomalyDetector membuat instance AnomalyDetector baru.
func NewAnomalyDetector(db *pgxpool.Pool, llmClient *ai.LLMClient) *AnomalyDetector {
	return &AnomalyDetector{db: db, llmClient: llmClient}
}

// Analyze menganalisis pola absensi karyawan dan mendeteksi anomali.
func (d *AnomalyDetector) Analyze(
	ctx context.Context,
	userID string,
	currentClockIn time.Time,
	gps GPSData,
	deviceHash string,
) ([]AnomalyResult, error) {
	pattern, err := d.getPattern(ctx, userID)
	if err != nil || pattern.SampleCount < 3 {
		// Tidak cukup data historis untuk analisis
		return nil, nil
	}

	var anomalies []AnomalyResult

	// Cek anomali waktu clock-in
	if timeAnomaly := d.checkTimeAnomaly(currentClockIn, pattern); timeAnomaly != nil {
		anomalies = append(anomalies, *timeAnomaly)
	}

	// Cek anomali lokasi
	if gps.Latitude != 0 && gps.Longitude != 0 {
		if locAnomaly := d.checkLocationAnomaly(gps, pattern); locAnomaly != nil {
			anomalies = append(anomalies, *locAnomaly)
		}
	}

	// Cek anomali device
	if deviceHash != "" && pattern.CommonDeviceHash != "" {
		if devAnomaly := d.checkDeviceAnomaly(deviceHash, pattern); devAnomaly != nil {
			anomalies = append(anomalies, *devAnomaly)
		}
	}

	// Jika ada anomali, minta AI untuk analisis lebih lanjut
	if len(anomalies) > 0 && d.llmClient != nil {
		aiAnalysis := d.analyzeWithAI(ctx, userID, anomalies, pattern, currentClockIn, gps)
		if aiAnalysis != "" {
			for i := range anomalies {
				anomalies[i].AIAnalysis = aiAnalysis
			}
		}
	}

	return anomalies, nil
}

// getPattern mengambil pola absensi karyawan dari 7 hari terakhir.
func (d *AnomalyDetector) getPattern(ctx context.Context, userID string) (*AttendancePattern, error) {
	query := `
		SELECT
			EXTRACT(HOUR FROM clock_in) + EXTRACT(MINUTE FROM clock_in) / 60.0 AS clock_in_hour,
			latitude,
			longitude,
			device_hash
		FROM attendances
		WHERE user_id = $1
		  AND clock_in IS NOT NULL
		  AND clock_in >= NOW() - INTERVAL '7 days'
		ORDER BY clock_in DESC
	`

	rows, err := d.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hours []float64
	var lats, lons []float64
	deviceCounts := make(map[string]int)

	for rows.Next() {
		var hour float64
		var lat, lon *float64
		var deviceHash *string

		if err := rows.Scan(&hour, &lat, &lon, &deviceHash); err != nil {
			continue
		}

		hours = append(hours, hour)
		if lat != nil && lon != nil {
			lats = append(lats, *lat)
			lons = append(lons, *lon)
		}
		if deviceHash != nil && *deviceHash != "" {
			deviceCounts[*deviceHash]++
		}
	}

	if len(hours) == 0 {
		return &AttendancePattern{SampleCount: 0}, nil
	}

	pattern := &AttendancePattern{
		SampleCount:     len(hours),
		AvgClockInHour:  average(hours),
		MedianLatitude:  median(lats),
		MedianLongitude: median(lons),
	}

	// Cari device yang paling sering dipakai
	maxCount := 0
	for hash, count := range deviceCounts {
		if count > maxCount {
			maxCount = count
			pattern.CommonDeviceHash = hash
		}
	}

	return pattern, nil
}

// checkTimeAnomaly memeriksa apakah waktu clock-in di luar jam normal ±2 jam.
func (d *AnomalyDetector) checkTimeAnomaly(clockIn time.Time, pattern *AttendancePattern) *AnomalyResult {
	currentHour := float64(clockIn.Hour()) + float64(clockIn.Minute())/60.0
	diff := math.Abs(currentHour - pattern.AvgClockInHour)

	// Normalisasi untuk melewati tengah malam
	if diff > 12 {
		diff = 24 - diff
	}

	if diff > 2 {
		return &AnomalyResult{
			HasAnomaly:  true,
			FraudType:   "anomaly_time",
			Severity:    "medium",
			Description: fmt.Sprintf("Clock-in di luar jam normal: %.1f jam dari rata-rata (%.2f vs %.2f)", diff, currentHour, pattern.AvgClockInHour),
			Evidence: Evidence{
				"current_hour":  currentHour,
				"average_hour":  pattern.AvgClockInHour,
				"diff_hours":    diff,
				"sample_count":  pattern.SampleCount,
			},
		}
	}

	return nil
}

// checkLocationAnomaly memeriksa apakah lokasi berbeda dari biasanya (> 5km dari median).
func (d *AnomalyDetector) checkLocationAnomaly(gps GPSData, pattern *AttendancePattern) *AnomalyResult {
	if pattern.MedianLatitude == 0 && pattern.MedianLongitude == 0 {
		return nil
	}

	distKm := haversineKm(pattern.MedianLatitude, pattern.MedianLongitude, gps.Latitude, gps.Longitude)

	if distKm > 5 {
		return &AnomalyResult{
			HasAnomaly:  true,
			FraudType:   "anomaly_location",
			Severity:    "medium",
			Description: fmt.Sprintf("Lokasi clock-in berbeda dari biasanya: %.1f km dari lokasi normal", distKm),
			Evidence: Evidence{
				"current_lat":    gps.Latitude,
				"current_lon":    gps.Longitude,
				"median_lat":     pattern.MedianLatitude,
				"median_lon":     pattern.MedianLongitude,
				"distance_km":    distKm,
				"sample_count":   pattern.SampleCount,
			},
		}
	}

	return nil
}

// checkDeviceAnomaly memeriksa apakah device berbeda dari yang biasa dipakai.
func (d *AnomalyDetector) checkDeviceAnomaly(deviceHash string, pattern *AttendancePattern) *AnomalyResult {
	if deviceHash == pattern.CommonDeviceHash {
		return nil
	}

	return &AnomalyResult{
		HasAnomaly:  true,
		FraudType:   "anomaly_device",
		Severity:    "low",
		Description: "Clock-in menggunakan device yang berbeda dari biasanya",
		Evidence: Evidence{
			"current_device":  deviceHash,
			"common_device":   pattern.CommonDeviceHash,
			"sample_count":    pattern.SampleCount,
		},
	}
}

// analyzeWithAI meminta AI untuk menganalisis pola anomali.
func (d *AnomalyDetector) analyzeWithAI(
	ctx context.Context,
	userID string,
	anomalies []AnomalyResult,
	pattern *AttendancePattern,
	clockIn time.Time,
	gps GPSData,
) string {
	anomalyDesc := ""
	for _, a := range anomalies {
		anomalyDesc += fmt.Sprintf("- %s: %s\n", a.FraudType, a.Description)
	}

	messages := []ai.ChatMessage{
		{
			Role:    "system",
			Content: "Anda adalah sistem deteksi fraud absensi karyawan. Berikan analisis singkat 2-3 kalimat.",
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`Analisis apakah pola absensi berikut mencurigakan:

Waktu clock-in saat ini: %s
Rata-rata jam clock-in normal: %.2f (%.0f:%.0f)
Lokasi saat ini: lat=%.6f, lon=%.6f
Anomali terdeteksi:
%s

Apakah ini kemungkinan fraud atau ada penjelasan wajar? Fokus pada konteks bisnis Indonesia.`,
				clockIn.Format("15:04 WIB"),
				pattern.AvgClockInHour,
				math.Floor(pattern.AvgClockInHour),
				math.Mod(pattern.AvgClockInHour, 1)*60,
				gps.Latitude, gps.Longitude,
				anomalyDesc,
			),
		},
	}

	response, err := d.llmClient.Chat(ctx, messages)
	if err != nil {
		return ""
	}

	return response
}

// average menghitung rata-rata slice float64.
func average(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := 0.0
	for _, n := range nums {
		sum += n
	}
	return sum / float64(len(nums))
}

// median menghitung nilai tengah slice float64.
func median(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	// Sorting sederhana untuk median
	sorted := make([]float64, len(nums))
	copy(sorted, nums)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}
