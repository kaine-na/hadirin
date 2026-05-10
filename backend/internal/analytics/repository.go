package analytics

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository menangani akses data analytics ke database.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat instance Repository baru.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetAttendanceSummary mengambil ringkasan kehadiran untuk periode tertentu.
func (r *Repository) GetAttendanceSummary(ctx context.Context, f AnalyticsFilter) (*AttendanceSummary, error) {
	query := `
		SELECT
			COUNT(DISTINCT u.id) AS total_employees,
			COUNT(a.id) AS total_records,
			COUNT(CASE WHEN a.status = 'hadir' THEN 1 END) AS total_present,
			COUNT(CASE WHEN a.status = 'terlambat' THEN 1 END) AS total_late,
			COUNT(CASE WHEN a.status = 'alpha' THEN 1 END) AS total_absent,
			COUNT(CASE WHEN a.status IN ('izin') THEN 1 END) AS total_leave,
			COUNT(CASE WHEN a.status = 'sakit' THEN 1 END) AS total_sick
		FROM users u
		LEFT JOIN attendances a ON a.user_id = u.id
			AND a.date BETWEEN $1::date AND $2::date
		WHERE u.role != 'super_admin'
	`
	args := []any{f.StartDate, f.EndDate}

	if f.DepartmentID != "" {
		query += " AND u.department = $3"
		args = append(args, f.DepartmentID)
	}

	s := &AttendanceSummary{
		PeriodStart: f.StartDate,
		PeriodEnd:   f.EndDate,
	}

	var totalRecords int
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&s.TotalEmployees,
		&totalRecords,
		&s.TotalPresent,
		&s.TotalLate,
		&s.TotalAbsent,
		&s.TotalLeave,
		&s.TotalSick,
	)
	if err != nil {
		return nil, fmt.Errorf("query attendance summary: %w", err)
	}

	s.TotalWorkingDays = totalRecords
	totalHadir := s.TotalPresent + s.TotalLate
	if totalRecords > 0 {
		s.AttendanceRate = float64(totalHadir) / float64(totalRecords) * 100
	}
	if totalHadir > 0 {
		s.LatenessRate = float64(s.TotalLate) / float64(totalHadir) * 100
	}

	return s, nil
}

// GetDepartmentStats mengambil statistik kehadiran per departemen.
func (r *Repository) GetDepartmentStats(ctx context.Context, f AnalyticsFilter) ([]*DepartmentStat, error) {
	query := `
		SELECT
			COALESCE(u.department, 'Tidak Ada Departemen') AS department,
			COUNT(DISTINCT u.id) AS total_employees,
			COUNT(CASE WHEN a.status = 'hadir' THEN 1 END) AS total_present,
			COUNT(CASE WHEN a.status = 'terlambat' THEN 1 END) AS total_late,
			COUNT(CASE WHEN a.status = 'alpha' THEN 1 END) AS total_absent,
			COUNT(a.id) AS total_records
		FROM users u
		LEFT JOIN attendances a ON a.user_id = u.id
			AND a.date BETWEEN $1::date AND $2::date
		WHERE u.role != 'super_admin'
		GROUP BY COALESCE(u.department, 'Tidak Ada Departemen')
		ORDER BY total_present DESC
	`

	rows, err := r.db.Query(ctx, query, f.StartDate, f.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query department stats: %w", err)
	}
	defer rows.Close()

	var stats []*DepartmentStat
	for rows.Next() {
		s := &DepartmentStat{}
		var totalRecords int
		if err := rows.Scan(
			&s.Department, &s.TotalEmployees,
			&s.TotalPresent, &s.TotalLate, &s.TotalAbsent,
			&totalRecords,
		); err != nil {
			return nil, fmt.Errorf("scan department stat: %w", err)
		}
		if totalRecords > 0 {
			s.AttendanceRate = float64(s.TotalPresent+s.TotalLate) / float64(totalRecords) * 100
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetTrend mengambil tren kehadiran harian untuk 30 hari terakhir.
func (r *Repository) GetTrend(ctx context.Context, f AnalyticsFilter) ([]*TrendPoint, error) {
	query := `
		SELECT
			a.date::text AS date,
			COUNT(CASE WHEN a.status = 'hadir' THEN 1 END) AS present,
			COUNT(CASE WHEN a.status = 'terlambat' THEN 1 END) AS late,
			COUNT(CASE WHEN a.status = 'alpha' THEN 1 END) AS absent
		FROM attendances a
		JOIN users u ON u.id = a.user_id
		WHERE a.date BETWEEN $1::date AND $2::date
			AND u.role != 'super_admin'
	`
	args := []any{f.StartDate, f.EndDate}

	if f.DepartmentID != "" {
		query += " AND u.department = $3"
		args = append(args, f.DepartmentID)
	}

	query += " GROUP BY a.date ORDER BY a.date ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query trend: %w", err)
	}
	defer rows.Close()

	var points []*TrendPoint
	for rows.Next() {
		p := &TrendPoint{}
		if err := rows.Scan(&p.Date, &p.Present, &p.Late, &p.Absent); err != nil {
			return nil, fmt.Errorf("scan trend point: %w", err)
		}
		points = append(points, p)
	}

	return points, nil
}

// GetTopLateEmployees mengambil top 10 karyawan paling sering terlambat.
func (r *Repository) GetTopLateEmployees(ctx context.Context, f AnalyticsFilter) ([]*TopLateEmployee, error) {
	query := `
		SELECT
			u.id::text,
			u.name,
			COALESCE(u.department, '-') AS department,
			COUNT(CASE WHEN a.status = 'terlambat' THEN 1 END) AS late_count,
			COUNT(CASE WHEN a.status = 'alpha' THEN 1 END) AS absent_count
		FROM users u
		JOIN attendances a ON a.user_id = u.id
			AND a.date BETWEEN $1::date AND $2::date
		WHERE u.role != 'super_admin'
	`
	args := []any{f.StartDate, f.EndDate}

	if f.DepartmentID != "" {
		query += " AND u.department = $3"
		args = append(args, f.DepartmentID)
	}

	query += `
		GROUP BY u.id, u.name, u.department
		HAVING COUNT(CASE WHEN a.status = 'terlambat' THEN 1 END) > 0
		ORDER BY late_count DESC
		LIMIT 10
	`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query top late employees: %w", err)
	}
	defer rows.Close()

	var employees []*TopLateEmployee
	for rows.Next() {
		e := &TopLateEmployee{}
		if err := rows.Scan(&e.EmployeeID, &e.Name, &e.Department, &e.LateCount, &e.AbsentCount); err != nil {
			return nil, fmt.Errorf("scan top late employee: %w", err)
		}
		employees = append(employees, e)
	}

	return employees, nil
}
