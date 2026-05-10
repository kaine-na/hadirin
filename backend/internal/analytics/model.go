package analytics

import "time"

// AttendanceSummary adalah ringkasan kehadiran untuk periode tertentu.
type AttendanceSummary struct {
	TotalEmployees    int     `json:"total_employees"`
	TotalWorkingDays  int     `json:"total_working_days"`
	TotalPresent      int     `json:"total_present"`
	TotalLate         int     `json:"total_late"`
	TotalAbsent       int     `json:"total_absent"`
	TotalLeave        int     `json:"total_leave"`
	TotalSick         int     `json:"total_sick"`
	AttendanceRate    float64 `json:"attendance_rate"`    // % hadir dari total hari kerja
	LatenessRate      float64 `json:"lateness_rate"`      // % terlambat dari total hadir
	PeriodStart       string  `json:"period_start"`
	PeriodEnd         string  `json:"period_end"`
}

// DepartmentStat adalah statistik kehadiran per departemen.
type DepartmentStat struct {
	Department     string  `json:"department"`
	TotalEmployees int     `json:"total_employees"`
	TotalPresent   int     `json:"total_present"`
	TotalLate      int     `json:"total_late"`
	TotalAbsent    int     `json:"total_absent"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// TrendPoint adalah satu titik data tren kehadiran harian.
type TrendPoint struct {
	Date    string `json:"date"`
	Present int    `json:"present"`
	Late    int    `json:"late"`
	Absent  int    `json:"absent"`
}

// TopLateEmployee adalah data karyawan yang paling sering terlambat.
type TopLateEmployee struct {
	EmployeeID string `json:"employee_id"`
	Name       string `json:"name"`
	Department string `json:"department"`
	LateCount  int    `json:"late_count"`
	AbsentCount int   `json:"absent_count"`
}

// AnalyticsFilter adalah parameter filter untuk query analytics.
type AnalyticsFilter struct {
	StartDate    string
	EndDate      string
	DepartmentID string
}

// ExecutiveSummaryRequest adalah request untuk generate AI summary.
type ExecutiveSummaryRequest struct {
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	DepartmentID string `json:"department_id,omitempty"`
}

// ExecutiveSummaryResponse adalah response AI executive summary.
type ExecutiveSummaryResponse struct {
	Summary     string    `json:"summary"`
	GeneratedAt time.Time `json:"generated_at"`
}
