package attendance

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository menangani semua operasi database untuk absensi.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat instance Repository baru.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FindTodayByUserID mengambil absensi hari ini untuk user tertentu.
func (r *Repository) FindTodayByUserID(ctx context.Context, userID string) (*Attendance, error) {
	att := &Attendance{}
	today := time.Now().Format("2006-01-02")

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id::text, date, clock_in, clock_out, status,
		       COALESCE(notes, ''), COALESCE(ip_address, ''),
		       COALESCE(created_by::text, ''), COALESCE(updated_by::text, ''),
		       created_at, updated_at
		FROM attendances
		WHERE user_id = $1 AND date = $2
	`, userID, today).Scan(
		&att.ID, &att.UserID, &att.Date, &att.ClockIn, &att.ClockOut, &att.Status,
		&att.Notes, &att.IPAddress, &att.CreatedBy, &att.UpdatedBy,
		&att.CreatedAt, &att.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return att, nil
}

// FindByID mengambil satu record absensi berdasarkan ID.
func (r *Repository) FindByID(ctx context.Context, id string) (*Attendance, error) {
	att := &Attendance{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id::text, date, clock_in, clock_out, status,
		       COALESCE(notes, ''), COALESCE(ip_address, ''),
		       COALESCE(created_by::text, ''), COALESCE(updated_by::text, ''),
		       created_at, updated_at
		FROM attendances WHERE id = $1
	`, id).Scan(
		&att.ID, &att.UserID, &att.Date, &att.ClockIn, &att.ClockOut, &att.Status,
		&att.Notes, &att.IPAddress, &att.CreatedBy, &att.UpdatedBy,
		&att.CreatedAt, &att.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("attendance not found: %w", err)
	}
	return att, nil
}

// ClockIn membuat record absensi baru (clock-in).
func (r *Repository) ClockIn(ctx context.Context, userID, ipAddress, notes string) (*Attendance, error) {
	att := &Attendance{}
	now := time.Now()
	today := now.Format("2006-01-02")

	// Tentukan status: terlambat jika setelah jam 09:00
	status := "hadir"
	if now.Hour() >= 9 {
		status = "terlambat"
	}

	err := r.db.QueryRow(ctx, `
		INSERT INTO attendances (user_id, date, clock_in, status, ip_address, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id::text, date, clock_in, clock_out, status,
		          COALESCE(notes, ''), COALESCE(ip_address, ''),
		          COALESCE(created_by::text, ''), COALESCE(updated_by::text, ''),
		          created_at, updated_at
	`, userID, today, now, status, ipAddress, nullableString(notes)).Scan(
		&att.ID, &att.UserID, &att.Date, &att.ClockIn, &att.ClockOut, &att.Status,
		&att.Notes, &att.IPAddress, &att.CreatedBy, &att.UpdatedBy,
		&att.CreatedAt, &att.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("clock in: %w", err)
	}
	return att, nil
}

// ClockOut mengupdate waktu clock-out.
func (r *Repository) ClockOut(ctx context.Context, attendanceID, notes string) (*Attendance, error) {
	att := &Attendance{}
	now := time.Now()

	err := r.db.QueryRow(ctx, `
		UPDATE attendances
		SET clock_out = $2, notes = COALESCE(NULLIF($3, ''), notes), updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id::text, date, clock_in, clock_out, status,
		          COALESCE(notes, ''), COALESCE(ip_address, ''),
		          COALESCE(created_by::text, ''), COALESCE(updated_by::text, ''),
		          created_at, updated_at
	`, attendanceID, now, notes).Scan(
		&att.ID, &att.UserID, &att.Date, &att.ClockIn, &att.ClockOut, &att.Status,
		&att.Notes, &att.IPAddress, &att.CreatedBy, &att.UpdatedBy,
		&att.CreatedAt, &att.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("clock out: %w", err)
	}
	return att, nil
}

// List mengambil daftar absensi dengan filter.
func (r *Repository) List(ctx context.Context, filter RekapFilter) ([]*Attendance, int, error) {
	args := []interface{}{filter.UserID}
	argIdx := 2
	conditions := "user_id = $1"

	if filter.StartDate != "" {
		conditions += fmt.Sprintf(" AND date >= $%d", argIdx)
		args = append(args, filter.StartDate)
		argIdx++
	}
	if filter.EndDate != "" {
		conditions += fmt.Sprintf(" AND date <= $%d", argIdx)
		args = append(args, filter.EndDate)
		argIdx++
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM attendances WHERE "+conditions, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count attendances: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, user_id::text, date, clock_in, clock_out, status,
		       COALESCE(notes, ''), COALESCE(ip_address, ''),
		       COALESCE(created_by::text, ''), COALESCE(updated_by::text, ''),
		       created_at, updated_at
		FROM attendances WHERE %s
		ORDER BY date DESC
		LIMIT $%d OFFSET $%d
	`, conditions, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list attendances: %w", err)
	}
	defer rows.Close()

	var attendances []*Attendance
	for rows.Next() {
		att := &Attendance{}
		if err := rows.Scan(
			&att.ID, &att.UserID, &att.Date, &att.ClockIn, &att.ClockOut, &att.Status,
			&att.Notes, &att.IPAddress, &att.CreatedBy, &att.UpdatedBy,
			&att.CreatedAt, &att.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan attendance: %w", err)
		}
		attendances = append(attendances, att)
	}

	return attendances, total, nil
}

// Override mengupdate absensi oleh HR.
func (r *Repository) Override(ctx context.Context, id, updatedBy string, req *OverrideRequest) (*Attendance, error) {
	att := &Attendance{}

	err := r.db.QueryRow(ctx, `
		UPDATE attendances SET
			status = COALESCE(NULLIF($2, ''), status),
			notes = COALESCE(NULLIF($3, ''), notes),
			clock_in = CASE WHEN $4 != '' THEN $4::timestamptz ELSE clock_in END,
			clock_out = CASE WHEN $5 != '' THEN $5::timestamptz ELSE clock_out END,
			updated_by = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id::text, date, clock_in, clock_out, status,
		          COALESCE(notes, ''), COALESCE(ip_address, ''),
		          COALESCE(created_by::text, ''), COALESCE(updated_by::text, ''),
		          created_at, updated_at
	`, id, req.Status, req.Notes, req.ClockIn, req.ClockOut, updatedBy).Scan(
		&att.ID, &att.UserID, &att.Date, &att.ClockIn, &att.ClockOut, &att.Status,
		&att.Notes, &att.IPAddress, &att.CreatedBy, &att.UpdatedBy,
		&att.CreatedAt, &att.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("override attendance: %w", err)
	}
	return att, nil
}

// ListForExport mengambil semua absensi untuk export CSV (tanpa pagination).
func (r *Repository) ListForExport(ctx context.Context, startDate, endDate string) ([]*Attendance, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.user_id::text, a.date, a.clock_in, a.clock_out, a.status,
		       COALESCE(a.notes, ''), COALESCE(a.ip_address, ''),
		       COALESCE(a.created_by::text, ''), COALESCE(a.updated_by::text, ''),
		       a.created_at, a.updated_at
		FROM attendances a
		WHERE ($1 = '' OR a.date >= $1::date)
		  AND ($2 = '' OR a.date <= $2::date)
		ORDER BY a.date DESC, a.user_id
	`, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("list for export: %w", err)
	}
	defer rows.Close()

	var attendances []*Attendance
	for rows.Next() {
		att := &Attendance{}
		if err := rows.Scan(
			&att.ID, &att.UserID, &att.Date, &att.ClockIn, &att.ClockOut, &att.Status,
			&att.Notes, &att.IPAddress, &att.CreatedBy, &att.UpdatedBy,
			&att.CreatedAt, &att.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan attendance: %w", err)
		}
		attendances = append(attendances, att)
	}

	return attendances, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
