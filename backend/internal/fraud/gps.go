package fraud

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GPSData berisi data lokasi dari request clock-in.
type GPSData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy"` // dalam meter
	UserAgent string  `json:"user_agent,omitempty"`
}

// GPSValidationResult hasil validasi GPS.
type GPSValidationResult struct {
	Valid       bool     `json:"valid"`
	Reason      string   `json:"reason,omitempty"`
	FraudType   string   `json:"fraud_type,omitempty"`
	Severity    string   `json:"severity,omitempty"`
	Description string   `json:"description,omitempty"`
	Evidence    Evidence `json:"evidence,omitempty"`
}

// Evidence adalah bukti fraud dalam format JSON.
type Evidence map[string]interface{}

// GPSValidator menangani validasi GPS untuk deteksi fraud.
type GPSValidator struct {
	db *pgxpool.Pool
}

// NewGPSValidator membuat instance GPSValidator baru.
func NewGPSValidator(db *pgxpool.Pool) *GPSValidator {
	return &GPSValidator{db: db}
}

// ValidateAccuracy memvalidasi akurasi GPS.
// Tolak jika accuracy > 100 meter (kemungkinan GPS palsu atau sinyal lemah).
func (v *GPSValidator) ValidateAccuracy(gps GPSData) *GPSValidationResult {
	if gps.Accuracy <= 0 {
		// Tidak ada data GPS — bukan error fatal, tapi catat
		return &GPSValidationResult{
			Valid:       true,
			Reason:      "no_gps_data",
			Description: "Data GPS tidak tersedia",
		}
	}

	if gps.Accuracy > 100 {
		return &GPSValidationResult{
			Valid:     false,
			FraudType: "gps_accuracy",
			Severity:  "low",
			Description: fmt.Sprintf(
				"Akurasi GPS terlalu rendah: %.1f meter (batas: 100 meter)",
				gps.Accuracy,
			),
			Evidence: Evidence{
				"accuracy":  gps.Accuracy,
				"threshold": 100,
				"latitude":  gps.Latitude,
				"longitude": gps.Longitude,
			},
		}
	}

	return &GPSValidationResult{Valid: true}
}

// DetectMockLocation mendeteksi kemungkinan mock location dari user-agent hints.
// Ini adalah deteksi sederhana berbasis heuristik.
func (v *GPSValidator) DetectMockLocation(gps GPSData) *GPSValidationResult {
	// Koordinat 0,0 adalah indikasi kuat mock location
	if gps.Latitude == 0 && gps.Longitude == 0 {
		return &GPSValidationResult{
			Valid:     false,
			FraudType: "mock_location",
			Severity:  "high",
			Description: "Koordinat GPS menunjukkan nilai 0,0 — kemungkinan mock location",
			Evidence: Evidence{
				"latitude":   gps.Latitude,
				"longitude":  gps.Longitude,
				"user_agent": gps.UserAgent,
			},
		}
	}

	// Koordinat di luar batas Indonesia (heuristik sederhana)
	// Indonesia: lat -11 s/d 6, lon 95 s/d 141
	if gps.Latitude != 0 && gps.Longitude != 0 {
		if gps.Latitude < -11 || gps.Latitude > 6 ||
			gps.Longitude < 95 || gps.Longitude > 141 {
			return &GPSValidationResult{
				Valid:     false,
				FraudType: "mock_location",
				Severity:  "medium",
				Description: fmt.Sprintf(
					"Koordinat GPS di luar wilayah Indonesia: lat=%.6f, lon=%.6f",
					gps.Latitude, gps.Longitude,
				),
				Evidence: Evidence{
					"latitude":  gps.Latitude,
					"longitude": gps.Longitude,
				},
			}
		}
	}

	return &GPSValidationResult{Valid: true}
}

// CheckVelocity memeriksa apakah karyawan melakukan clock-in dari 2 lokasi
// yang terlalu jauh dalam waktu singkat (velocity check).
// Flag jika: 2 lokasi > 50km dalam 30 menit.
func (v *GPSValidator) CheckVelocity(ctx context.Context, userID string, gps GPSData) (*GPSValidationResult, error) {
	if gps.Latitude == 0 && gps.Longitude == 0 {
		return &GPSValidationResult{Valid: true}, nil
	}

	// Ambil clock-in terakhir dalam 30 menit terakhir
	query := `
		SELECT latitude, longitude, clock_in
		FROM attendances
		WHERE user_id = $1
		  AND clock_in IS NOT NULL
		  AND clock_in >= NOW() - INTERVAL '30 minutes'
		  AND latitude IS NOT NULL
		  AND longitude IS NOT NULL
		ORDER BY clock_in DESC
		LIMIT 1
	`

	var prevLat, prevLon float64
	var prevClockIn time.Time

	err := v.db.QueryRow(ctx, query, userID).Scan(&prevLat, &prevLon, &prevClockIn)
	if err != nil {
		// Tidak ada data sebelumnya — tidak ada masalah
		return &GPSValidationResult{Valid: true}, nil
	}

	// Hitung jarak menggunakan Haversine formula
	distKm := haversineKm(prevLat, prevLon, gps.Latitude, gps.Longitude)
	timeDiff := time.Since(prevClockIn).Minutes()

	if distKm > 50 {
		return &GPSValidationResult{
			Valid:     false,
			FraudType: "velocity_check",
			Severity:  "critical",
			Description: fmt.Sprintf(
				"Velocity anomaly: clock-in dari 2 lokasi berjarak %.1f km dalam %.0f menit",
				distKm, timeDiff,
			),
			Evidence: Evidence{
				"previous_lat":    prevLat,
				"previous_lon":    prevLon,
				"previous_time":   prevClockIn.Format(time.RFC3339),
				"current_lat":     gps.Latitude,
				"current_lon":     gps.Longitude,
				"distance_km":     distKm,
				"time_diff_minutes": timeDiff,
			},
		}, nil
	}

	return &GPSValidationResult{Valid: true}, nil
}

// haversineKm menghitung jarak antara 2 koordinat dalam kilometer.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}
