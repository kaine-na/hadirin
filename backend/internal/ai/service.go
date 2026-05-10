package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service menangani logika bisnis untuk modul AI.
type Service struct {
	db     *pgxpool.Pool
	client *LLMClient
	model  string
}

// NewService membuat instance Service baru.
func NewService(db *pgxpool.Pool, client *LLMClient, model string) *Service {
	return &Service{db: db, client: client, model: model}
}

// Analyze menghasilkan laporan AI untuk satu karyawan.
func (s *Service) Analyze(ctx context.Context, employeeID, generatedBy string, req *AnalyzeRequest) (*AIReport, error) {
	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		return nil, fmt.Errorf("format period_start tidak valid (gunakan YYYY-MM-DD): %w", err)
	}
	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("format period_end tidak valid (gunakan YYYY-MM-DD): %w", err)
	}

	employeeData, err := s.getEmployeeData(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("ambil data karyawan: %w", err)
	}

	attendanceData, err := s.getAttendanceData(ctx, employeeID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("ambil data absensi: %w", err)
	}

	documentData, err := s.getDocumentData(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("ambil data dokumen: %w", err)
	}

	prompt := s.buildPrompt(employeeData, attendanceData, documentData, req.PeriodStart, req.PeriodEnd, req.CustomPrompt)

	messages := []ChatMessage{
		{
			Role:    "system",
			Content: "Kamu adalah asisten HR yang membantu menganalisis performa karyawan berdasarkan data absensi dan dokumen. Berikan analisis yang objektif, konstruktif, dan dalam Bahasa Indonesia.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	llmResponse, err := s.client.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("panggil LLM: %w", err)
	}

	report := &AIReport{
		EmployeeID:  employeeID,
		GeneratedBy: generatedBy,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Prompt:      prompt,
		Response:    llmResponse,
		ModelUsed:   s.model,
	}

	if err := s.saveReport(ctx, report); err != nil {
		return nil, fmt.Errorf("simpan laporan: %w", err)
	}

	return report, nil
}

// GetReports mengambil riwayat laporan AI untuk satu karyawan.
func (s *Service) GetReports(ctx context.Context, employeeID string) ([]*AIReport, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, employee_id::text, generated_by::text, period_start, period_end,
		       prompt, response, COALESCE(model_used, ''), created_at
		FROM ai_reports
		WHERE employee_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	var reports []*AIReport
	for rows.Next() {
		r := &AIReport{}
		if err := rows.Scan(
			&r.ID, &r.EmployeeID, &r.GeneratedBy, &r.PeriodStart, &r.PeriodEnd,
			&r.Prompt, &r.Response, &r.ModelUsed, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		reports = append(reports, r)
	}

	return reports, nil
}

// GetReportByID mengambil satu laporan AI berdasarkan ID.
func (s *Service) GetReportByID(ctx context.Context, id string) (*AIReport, error) {
	r := &AIReport{}
	err := s.db.QueryRow(ctx, `
		SELECT id, employee_id::text, generated_by::text, period_start, period_end,
		       prompt, response, COALESCE(model_used, ''), created_at
		FROM ai_reports WHERE id = $1
	`, id).Scan(
		&r.ID, &r.EmployeeID, &r.GeneratedBy, &r.PeriodStart, &r.PeriodEnd,
		&r.Prompt, &r.Response, &r.ModelUsed, &r.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("report not found: %w", err)
	}
	return r, nil
}

func (s *Service) getEmployeeData(ctx context.Context, employeeID string) (string, error) {
	var name, email, role, department, position string
	err := s.db.QueryRow(ctx, `
		SELECT name, email, role, COALESCE(department, '-'), COALESCE(position, '-')
		FROM users WHERE id = $1
	`, employeeID).Scan(&name, &email, &role, &department, &position)
	if err != nil {
		return "", fmt.Errorf("employee not found: %w", err)
	}
	return fmt.Sprintf("Nama: %s | Email: %s | Role: %s | Departemen: %s | Jabatan: %s",
		name, email, role, department, position), nil
}

func (s *Service) getAttendanceData(ctx context.Context, employeeID, startDate, endDate string) (string, error) {
	rows, err := s.db.Query(ctx, `
		SELECT date, status, clock_in, clock_out, COALESCE(notes, '')
		FROM attendances
		WHERE user_id = $1 AND date BETWEEN $2::date AND $3::date
		ORDER BY date
	`, employeeID, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("query attendance: %w", err)
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("Data Absensi:\n")

	var hadir, terlambat, izin, sakit, alpha int
	for rows.Next() {
		var date time.Time
		var status, notes string
		var clockIn, clockOut *time.Time
		if err := rows.Scan(&date, &status, &clockIn, &clockOut, &notes); err != nil {
			continue
		}
		switch status {
		case "hadir":
			hadir++
		case "terlambat":
			terlambat++
		case "izin":
			izin++
		case "sakit":
			sakit++
		case "alpha":
			alpha++
		}
	}

	sb.WriteString(fmt.Sprintf("- Hadir: %d hari\n", hadir))
	sb.WriteString(fmt.Sprintf("- Terlambat: %d hari\n", terlambat))
	sb.WriteString(fmt.Sprintf("- Izin: %d hari\n", izin))
	sb.WriteString(fmt.Sprintf("- Sakit: %d hari\n", sakit))
	sb.WriteString(fmt.Sprintf("- Alpha: %d hari\n", alpha))

	return sb.String(), nil
}

func (s *Service) getDocumentData(ctx context.Context, employeeID string) (string, error) {
	rows, err := s.db.Query(ctx, `
		SELECT title, category, created_at
		FROM documents
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 20
	`, employeeID)
	if err != nil {
		return "", fmt.Errorf("query documents: %w", err)
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("Dokumen yang Diupload:\n")

	count := 0
	for rows.Next() {
		var title, category string
		var createdAt time.Time
		if err := rows.Scan(&title, &category, &createdAt); err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", category, title, createdAt.Format("2006-01-02")))
		count++
	}

	if count == 0 {
		sb.WriteString("- Belum ada dokumen\n")
	}

	return sb.String(), nil
}

func (s *Service) buildPrompt(employeeData, attendanceData, documentData, startDate, endDate, customPrompt string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Analisis performa karyawan untuk periode %s sampai %s.\n\n", startDate, endDate))
	sb.WriteString("=== DATA KARYAWAN ===\n")
	sb.WriteString(employeeData + "\n\n")
	sb.WriteString("=== DATA ABSENSI ===\n")
	sb.WriteString(attendanceData + "\n")
	sb.WriteString("=== DOKUMEN ===\n")
	sb.WriteString(documentData + "\n")

	if customPrompt != "" {
		sb.WriteString("\n=== INSTRUKSI KHUSUS ===\n")
		sb.WriteString(customPrompt + "\n")
	} else {
		sb.WriteString("\nBerikan analisis yang mencakup:\n")
		sb.WriteString("1. Ringkasan kehadiran dan kedisiplinan\n")
		sb.WriteString("2. Pola yang perlu diperhatikan\n")
		sb.WriteString("3. Rekomendasi untuk karyawan dan HR\n")
	}

	return sb.String()
}

func (s *Service) saveReport(ctx context.Context, report *AIReport) error {
	return s.db.QueryRow(ctx, `
		INSERT INTO ai_reports (employee_id, generated_by, period_start, period_end, prompt, response, model_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`,
		report.EmployeeID, report.GeneratedBy, report.PeriodStart, report.PeriodEnd,
		report.Prompt, report.Response, report.ModelUsed,
	).Scan(&report.ID, &report.CreatedAt)
}
