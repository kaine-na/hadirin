package fraud

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FraudLog adalah representasi log fraud dari database.
type FraudLog struct {
	ID           string     `json:"id"`
	AttendanceID *string    `json:"attendance_id,omitempty"`
	UserID       string     `json:"user_id"`
	EmployeeName string     `json:"employee_name,omitempty"`
	FraudType    string     `json:"fraud_type"`
	Severity     string     `json:"severity"`
	Description  string     `json:"description"`
	Evidence     Evidence   `json:"evidence,omitempty"`
	Status       string     `json:"status"`
	ReviewedBy   *string    `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes  string     `json:"review_notes,omitempty"`
	AIAnalysis   string     `json:"ai_analysis,omitempty"`
	AIConfidence *float64   `json:"ai_confidence,omitempty"`
	PhotoURL     string     `json:"photo_url,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// FraudSummary ringkasan fraud bulan ini.
type FraudSummary struct {
	TotalLogs      int            `json:"total_logs"`
	PendingLogs    int            `json:"pending_logs"`
	ConfirmedLogs  int            `json:"confirmed_logs"`
	DismissedLogs  int            `json:"dismissed_logs"`
	ByType         map[string]int `json:"by_type"`
	BySeverity     map[string]int `json:"by_severity"`
	TopEmployees   []TopEmployee  `json:"top_employees"`
}

// TopEmployee karyawan dengan fraud terbanyak.
type TopEmployee struct {
	UserID       string `json:"user_id"`
	EmployeeName string `json:"employee_name"`
	FraudCount   int    `json:"fraud_count"`
}

// ReviewRequest request untuk review fraud log.
type ReviewRequest struct {
	Notes string `json:"notes"`
}

// Repository menangani operasi database untuk fraud detection.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat instance Repository baru.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateFraudLog menyimpan log fraud baru ke database.
func (r *Repository) CreateFraudLog(ctx context.Context, log *FraudLog) (string, error) {
	evidenceJSON, err := json.Marshal(log.Evidence)
	if err != nil {
		evidenceJSON = []byte("{}")
	}

	query := `
		INSERT INTO fraud_logs (
			attendance_id, user_id, fraud_type, severity,
			description, evidence, status, ai_analysis, ai_confidence
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id string
	err = r.db.QueryRow(ctx, query,
		log.AttendanceID, log.UserID, log.FraudType, log.Severity,
		log.Description, evidenceJSON, "pending", log.AIAnalysis, log.AIConfidence,
	).Scan(&id)

	return id, err
}

// ListFraudLogs mengambil daftar fraud logs dengan filter.
func (r *Repository) ListFraudLogs(ctx context.Context, status string, page, pageSize int) ([]FraudLog, int, error) {
	offset := (page - 1) * pageSize

	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		whereClause = fmt.Sprintf("WHERE fl.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM fraud_logs fl %s`, whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT
			fl.id, fl.attendance_id, fl.user_id,
			COALESCE(u.name, '') AS employee_name,
			fl.fraud_type, fl.severity, fl.description,
			fl.evidence, fl.status,
			fl.reviewed_by, fl.reviewed_at, fl.review_notes,
			fl.ai_analysis, fl.ai_confidence,
			fl.created_at, fl.updated_at,
			COALESCE(
				(SELECT file_path FROM attendance_photos WHERE attendance_id = fl.attendance_id LIMIT 1),
				''
			) AS photo_path
		FROM fraud_logs fl
		LEFT JOIN users u ON u.id = fl.user_id
		%s
		ORDER BY fl.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []FraudLog
	for rows.Next() {
		var log FraudLog
		var evidenceJSON []byte
		var photoPath string

		err := rows.Scan(
			&log.ID, &log.AttendanceID, &log.UserID, &log.EmployeeName,
			&log.FraudType, &log.Severity, &log.Description,
			&evidenceJSON, &log.Status,
			&log.ReviewedBy, &log.ReviewedAt, &log.ReviewNotes,
			&log.AIAnalysis, &log.AIConfidence,
			&log.CreatedAt, &log.UpdatedAt,
			&photoPath,
		)
		if err != nil {
			continue
		}

		if len(evidenceJSON) > 0 {
			_ = json.Unmarshal(evidenceJSON, &log.Evidence)
		}

		if photoPath != "" {
			log.PhotoURL = "/uploads/" + photoPath
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetFraudLogByID mengambil detail fraud log berdasarkan ID.
func (r *Repository) GetFraudLogByID(ctx context.Context, id string) (*FraudLog, error) {
	query := `
		SELECT
			fl.id, fl.attendance_id, fl.user_id,
			COALESCE(u.name, '') AS employee_name,
			fl.fraud_type, fl.severity, fl.description,
			fl.evidence, fl.status,
			fl.reviewed_by, fl.reviewed_at, fl.review_notes,
			fl.ai_analysis, fl.ai_confidence,
			fl.created_at, fl.updated_at,
			COALESCE(
				(SELECT file_path FROM attendance_photos WHERE attendance_id = fl.attendance_id LIMIT 1),
				''
			) AS photo_path
		FROM fraud_logs fl
		LEFT JOIN users u ON u.id = fl.user_id
		WHERE fl.id = $1
	`

	var log FraudLog
	var evidenceJSON []byte
	var photoPath string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&log.ID, &log.AttendanceID, &log.UserID, &log.EmployeeName,
		&log.FraudType, &log.Severity, &log.Description,
		&evidenceJSON, &log.Status,
		&log.ReviewedBy, &log.ReviewedAt, &log.ReviewNotes,
		&log.AIAnalysis, &log.AIConfidence,
		&log.CreatedAt, &log.UpdatedAt,
		&photoPath,
	)
	if err != nil {
		return nil, err
	}

	if len(evidenceJSON) > 0 {
		_ = json.Unmarshal(evidenceJSON, &log.Evidence)
	}

	if photoPath != "" {
		log.PhotoURL = "/uploads/" + photoPath
	}

	return &log, nil
}

// UpdateFraudLogStatus mengupdate status fraud log (dismiss/confirm).
func (r *Repository) UpdateFraudLogStatus(ctx context.Context, id, status, reviewerID, notes string) error {
	query := `
		UPDATE fraud_logs
		SET status = $1, reviewed_by = $2, reviewed_at = NOW(), review_notes = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, status, reviewerID, notes, id)
	return err
}

// GetFraudSummary mengambil ringkasan fraud bulan ini.
func (r *Repository) GetFraudSummary(ctx context.Context) (*FraudSummary, error) {
	summary := &FraudSummary{
		ByType:     make(map[string]int),
		BySeverity: make(map[string]int),
	}

	// Total dan per status
	statusQuery := `
		SELECT status, COUNT(*) FROM fraud_logs
		WHERE created_at >= date_trunc('month', NOW())
		GROUP BY status
	`
	rows, err := r.db.Query(ctx, statusQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		summary.TotalLogs += count
		switch status {
		case "pending":
			summary.PendingLogs = count
		case "confirmed":
			summary.ConfirmedLogs = count
		case "dismissed":
			summary.DismissedLogs = count
		}
	}

	// Per tipe fraud
	typeQuery := `
		SELECT fraud_type, COUNT(*) FROM fraud_logs
		WHERE created_at >= date_trunc('month', NOW())
		GROUP BY fraud_type
	`
	typeRows, err := r.db.Query(ctx, typeQuery)
	if err != nil {
		return nil, err
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var fraudType string
		var count int
		if err := typeRows.Scan(&fraudType, &count); err != nil {
			continue
		}
		summary.ByType[fraudType] = count
	}

	// Per severity
	sevQuery := `
		SELECT severity, COUNT(*) FROM fraud_logs
		WHERE created_at >= date_trunc('month', NOW())
		GROUP BY severity
	`
	sevRows, err := r.db.Query(ctx, sevQuery)
	if err != nil {
		return nil, err
	}
	defer sevRows.Close()

	for sevRows.Next() {
		var severity string
		var count int
		if err := sevRows.Scan(&severity, &count); err != nil {
			continue
		}
		summary.BySeverity[severity] = count
	}

	// Top employees dengan fraud terbanyak
	topQuery := `
		SELECT fl.user_id, COALESCE(u.name, 'Unknown'), COUNT(*) AS fraud_count
		FROM fraud_logs fl
		LEFT JOIN users u ON u.id = fl.user_id
		WHERE fl.created_at >= date_trunc('month', NOW())
		  AND fl.status != 'dismissed'
		GROUP BY fl.user_id, u.name
		ORDER BY fraud_count DESC
		LIMIT 5
	`
	topRows, err := r.db.Query(ctx, topQuery)
	if err != nil {
		return nil, err
	}
	defer topRows.Close()

	for topRows.Next() {
		var emp TopEmployee
		if err := topRows.Scan(&emp.UserID, &emp.EmployeeName, &emp.FraudCount); err != nil {
			continue
		}
		summary.TopEmployees = append(summary.TopEmployees, emp)
	}

	return summary, nil
}

// UpsertDeviceFingerprint menyimpan atau mengupdate fingerprint device.
func (r *Repository) UpsertDeviceFingerprint(ctx context.Context, userID, deviceHash, userAgent string) error {
	query := `
		INSERT INTO device_fingerprints (user_id, device_hash, user_agent)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, device_hash)
		DO UPDATE SET last_seen = NOW(), user_agent = EXCLUDED.user_agent
	`
	_, err := r.db.Exec(ctx, query, userID, deviceHash, userAgent)
	return err
}
