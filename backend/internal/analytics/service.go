package analytics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"saas-karyawan/internal/ai"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service menangani logika bisnis analytics.
type Service struct {
	repo      *Repository
	llmClient *ai.LLMClient
	model     string
}

// NewService membuat instance Service baru.
func NewService(db *pgxpool.Pool, llmClient *ai.LLMClient, model string) *Service {
	return &Service{
		repo:      NewRepository(db),
		llmClient: llmClient,
		model:     model,
	}
}

// GetAttendanceSummary mengambil ringkasan kehadiran.
func (s *Service) GetAttendanceSummary(ctx context.Context, f AnalyticsFilter) (*AttendanceSummary, error) {
	return s.repo.GetAttendanceSummary(ctx, f)
}

// GetDepartmentStats mengambil statistik per departemen.
func (s *Service) GetDepartmentStats(ctx context.Context, f AnalyticsFilter) ([]*DepartmentStat, error) {
	return s.repo.GetDepartmentStats(ctx, f)
}

// GetTrend mengambil tren kehadiran harian.
func (s *Service) GetTrend(ctx context.Context, f AnalyticsFilter) ([]*TrendPoint, error) {
	return s.repo.GetTrend(ctx, f)
}

// GetTopLateEmployees mengambil top 10 karyawan paling sering terlambat.
func (s *Service) GetTopLateEmployees(ctx context.Context, f AnalyticsFilter) ([]*TopLateEmployee, error) {
	return s.repo.GetTopLateEmployees(ctx, f)
}

// GenerateExecutiveSummary menghasilkan ringkasan eksekutif AI dari data analytics.
func (s *Service) GenerateExecutiveSummary(ctx context.Context, f AnalyticsFilter) (*ExecutiveSummaryResponse, error) {
	// Ambil semua data analytics
	summary, err := s.repo.GetAttendanceSummary(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("ambil summary: %w", err)
	}

	deptStats, err := s.repo.GetDepartmentStats(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("ambil dept stats: %w", err)
	}

	topLate, err := s.repo.GetTopLateEmployees(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("ambil top late: %w", err)
	}

	// Bangun prompt
	prompt := s.buildSummaryPrompt(summary, deptStats, topLate)

	messages := []ai.ChatMessage{
		{
			Role:    "system",
			Content: "Kamu adalah analis HR senior yang membuat ringkasan eksekutif laporan kehadiran karyawan. Tulis dalam Bahasa Indonesia yang profesional dan ringkas.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	text, err := s.llmClient.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("panggil LLM: %w", err)
	}

	return &ExecutiveSummaryResponse{
		Summary:     text,
		GeneratedAt: time.Now(),
	}, nil
}

func (s *Service) buildSummaryPrompt(
	summary *AttendanceSummary,
	deptStats []*DepartmentStat,
	topLate []*TopLateEmployee,
) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(
		"Buat ringkasan eksekutif 3 paragraf dari data HR berikut untuk periode %s sampai %s:\n\n",
		summary.PeriodStart, summary.PeriodEnd,
	))

	sb.WriteString("=== RINGKASAN KEHADIRAN ===\n")
	sb.WriteString(fmt.Sprintf("- Total karyawan: %d\n", summary.TotalEmployees))
	sb.WriteString(fmt.Sprintf("- Total hadir: %d hari\n", summary.TotalPresent))
	sb.WriteString(fmt.Sprintf("- Total terlambat: %d hari\n", summary.TotalLate))
	sb.WriteString(fmt.Sprintf("- Total alpha: %d hari\n", summary.TotalAbsent))
	sb.WriteString(fmt.Sprintf("- Total izin: %d hari\n", summary.TotalLeave))
	sb.WriteString(fmt.Sprintf("- Total sakit: %d hari\n", summary.TotalSick))
	sb.WriteString(fmt.Sprintf("- Tingkat kehadiran: %.1f%%\n", summary.AttendanceRate))
	sb.WriteString(fmt.Sprintf("- Tingkat keterlambatan: %.1f%%\n\n", summary.LatenessRate))

	sb.WriteString("=== STATISTIK PER DEPARTEMEN ===\n")
	for _, d := range deptStats {
		sb.WriteString(fmt.Sprintf("- %s: %d karyawan, kehadiran %.1f%%\n",
			d.Department, d.TotalEmployees, d.AttendanceRate))
	}

	if len(topLate) > 0 {
		sb.WriteString("\n=== TOP KARYAWAN TERLAMBAT ===\n")
		for i, e := range topLate {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("- %s (%s): %d kali terlambat\n",
				e.Name, e.Department, e.LateCount))
		}
	}

	sb.WriteString("\nFokus pada: tren kehadiran, departemen yang perlu perhatian, dan rekomendasi tindakan HR.")

	return sb.String()
}
