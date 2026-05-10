package leave

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository menangani semua operasi database untuk modul leave.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat instance Repository baru.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetLeaveTypes mengambil semua jenis cuti.
func (r *Repository) GetLeaveTypes(ctx context.Context) ([]*LeaveType, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, max_days, is_paid, COALESCE(description, ''), created_at, updated_at
		FROM leave_types
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("query leave types: %w", err)
	}
	defer rows.Close()

	var types []*LeaveType
	for rows.Next() {
		lt := &LeaveType{}
		if err := rows.Scan(&lt.ID, &lt.Name, &lt.MaxDays, &lt.IsPaid, &lt.Description, &lt.CreatedAt, &lt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan leave type: %w", err)
		}
		types = append(types, lt)
	}
	return types, nil
}

// CreateLeaveRequest membuat pengajuan cuti baru.
func (r *Repository) CreateLeaveRequest(ctx context.Context, req *LeaveRequest) (*LeaveRequest, error) {
	result := &LeaveRequest{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO leave_requests (user_id, leave_type_id, start_date, end_date, total_days, reason, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
		RETURNING id, user_id::text, leave_type_id::text, start_date, end_date, total_days,
		          reason, status, COALESCE(approved_by::text, ''), approved_at,
		          COALESCE(rejection_reason, ''), created_at, updated_at
	`, req.UserID, req.LeaveTypeID, req.StartDate, req.EndDate, req.TotalDays, req.Reason).Scan(
		&result.ID, &result.UserID, &result.LeaveTypeID, &result.StartDate, &result.EndDate,
		&result.TotalDays, &result.Reason, &result.Status, &result.ApprovedBy,
		&result.ApprovedAt, &result.RejectionReason, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create leave request: %w", err)
	}
	return result, nil
}

// FindByID mengambil satu pengajuan cuti berdasarkan ID.
func (r *Repository) FindByID(ctx context.Context, id string) (*LeaveRequest, error) {
	req := &LeaveRequest{}
	err := r.db.QueryRow(ctx, `
		SELECT lr.id, lr.user_id::text, COALESCE(u.name, ''),
		       lr.leave_type_id::text, COALESCE(lt.name, ''),
		       lr.start_date, lr.end_date, lr.total_days, lr.reason, lr.status,
		       COALESCE(lr.approved_by::text, ''), COALESCE(ab.name, ''),
		       lr.approved_at, COALESCE(lr.rejection_reason, ''),
		       lr.created_at, lr.updated_at
		FROM leave_requests lr
		LEFT JOIN users u ON u.id = lr.user_id
		LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id
		LEFT JOIN users ab ON ab.id = lr.approved_by
		WHERE lr.id = $1
	`, id).Scan(
		&req.ID, &req.UserID, &req.UserName,
		&req.LeaveTypeID, &req.LeaveTypeName,
		&req.StartDate, &req.EndDate, &req.TotalDays, &req.Reason, &req.Status,
		&req.ApprovedBy, &req.ApprovedByName,
		&req.ApprovedAt, &req.RejectionReason,
		&req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("leave request not found: %w", err)
	}
	return req, nil
}

// List mengambil daftar pengajuan cuti dengan filter.
func (r *Repository) List(ctx context.Context, filter LeaveFilter) ([]*LeaveRequest, int, error) {
	args := []interface{}{}
	argIdx := 1
	conditions := "1=1"

	// HR melihat semua, karyawan hanya milik sendiri
	if filter.UserID != "" {
		conditions += fmt.Sprintf(" AND lr.user_id = $%d", argIdx)
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Status != "" {
		conditions += fmt.Sprintf(" AND lr.status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.StartDate != "" {
		conditions += fmt.Sprintf(" AND lr.start_date >= $%d", argIdx)
		args = append(args, filter.StartDate)
		argIdx++
	}
	if filter.EndDate != "" {
		conditions += fmt.Sprintf(" AND lr.end_date <= $%d", argIdx)
		args = append(args, filter.EndDate)
		argIdx++
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM leave_requests lr WHERE "+conditions, countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count leave requests: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT lr.id, lr.user_id::text, COALESCE(u.name, ''),
		       lr.leave_type_id::text, COALESCE(lt.name, ''),
		       lr.start_date, lr.end_date, lr.total_days, lr.reason, lr.status,
		       COALESCE(lr.approved_by::text, ''), COALESCE(ab.name, ''),
		       lr.approved_at, COALESCE(lr.rejection_reason, ''),
		       lr.created_at, lr.updated_at
		FROM leave_requests lr
		LEFT JOIN users u ON u.id = lr.user_id
		LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id
		LEFT JOIN users ab ON ab.id = lr.approved_by
		WHERE %s
		ORDER BY lr.created_at DESC
		LIMIT $%d OFFSET $%d
	`, conditions, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list leave requests: %w", err)
	}
	defer rows.Close()

	var requests []*LeaveRequest
	for rows.Next() {
		req := &LeaveRequest{}
		if err := rows.Scan(
			&req.ID, &req.UserID, &req.UserName,
			&req.LeaveTypeID, &req.LeaveTypeName,
			&req.StartDate, &req.EndDate, &req.TotalDays, &req.Reason, &req.Status,
			&req.ApprovedBy, &req.ApprovedByName,
			&req.ApprovedAt, &req.RejectionReason,
			&req.CreatedAt, &req.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan leave request: %w", err)
		}
		requests = append(requests, req)
	}
	return requests, total, nil
}

// UpdateStatus mengupdate status pengajuan cuti.
func (r *Repository) UpdateStatus(ctx context.Context, id, approvedBy string, status LeaveStatus, rejectionReason string) (*LeaveRequest, error) {
	var approvedAt *time.Time
	if status == StatusApproved || status == StatusRejected {
		now := time.Now()
		approvedAt = &now
	}

	req := &LeaveRequest{}
	err := r.db.QueryRow(ctx, `
		UPDATE leave_requests SET
			status = $2,
			approved_by = CASE WHEN $3 != '' THEN $3::uuid ELSE approved_by END,
			approved_at = $4,
			rejection_reason = CASE WHEN $5 != '' THEN $5 ELSE rejection_reason END,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id::text, leave_type_id::text, start_date, end_date, total_days,
		          reason, status, COALESCE(approved_by::text, ''), approved_at,
		          COALESCE(rejection_reason, ''), created_at, updated_at
	`, id, status, approvedBy, approvedAt, rejectionReason).Scan(
		&req.ID, &req.UserID, &req.LeaveTypeID, &req.StartDate, &req.EndDate,
		&req.TotalDays, &req.Reason, &req.Status, &req.ApprovedBy,
		&req.ApprovedAt, &req.RejectionReason, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update leave status: %w", err)
	}
	return req, nil
}

// GetLeaveBalance mengambil saldo cuti user untuk tahun tertentu.
func (r *Repository) GetLeaveBalance(ctx context.Context, userID string, year int) ([]*LeaveBalance, error) {
	rows, err := r.db.Query(ctx, `
		SELECT lb.id, lb.user_id::text, lb.leave_type_id::text, lt.name,
		       lb.year, lb.total_days, lb.used_days, lb.remaining_days,
		       lb.created_at, lb.updated_at
		FROM leave_balances lb
		JOIN leave_types lt ON lt.id = lb.leave_type_id
		WHERE lb.user_id = $1 AND lb.year = $2
		ORDER BY lt.name
	`, userID, year)
	if err != nil {
		return nil, fmt.Errorf("query leave balance: %w", err)
	}
	defer rows.Close()

	var balances []*LeaveBalance
	for rows.Next() {
		b := &LeaveBalance{}
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.LeaveTypeID, &b.LeaveTypeName,
			&b.Year, &b.TotalDays, &b.UsedDays, &b.RemainingDays,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan leave balance: %w", err)
		}
		balances = append(balances, b)
	}
	return balances, nil
}

// UpdateLeaveBalance mengurangi saldo cuti saat pengajuan disetujui.
func (r *Repository) UpdateLeaveBalance(ctx context.Context, userID, leaveTypeID string, year, days int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO leave_balances (user_id, leave_type_id, year, total_days, used_days)
		VALUES ($1, $2, $3,
			COALESCE((SELECT max_days FROM leave_types WHERE id = $2), 0),
			$4
		)
		ON CONFLICT (user_id, leave_type_id, year)
		DO UPDATE SET
			used_days = leave_balances.used_days + $4,
			updated_at = NOW()
	`, userID, leaveTypeID, year, days)
	if err != nil {
		return fmt.Errorf("update leave balance: %w", err)
	}
	return nil
}

// RestoreLeaveBalance mengembalikan saldo cuti saat pengajuan dibatalkan/ditolak.
func (r *Repository) RestoreLeaveBalance(ctx context.Context, userID, leaveTypeID string, year, days int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE leave_balances
		SET used_days = GREATEST(0, used_days - $4), updated_at = NOW()
		WHERE user_id = $1 AND leave_type_id = $2 AND year = $3
	`, userID, leaveTypeID, year, days)
	if err != nil {
		return fmt.Errorf("restore leave balance: %w", err)
	}
	return nil
}

// GetTeamLeaveHistory mengambil histori cuti tim dalam N bulan terakhir untuk AI recommendation.
func (r *Repository) GetTeamLeaveHistory(ctx context.Context, months int) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.name, lt.name as leave_type, lr.start_date, lr.end_date,
		       lr.total_days, lr.status
		FROM leave_requests lr
		JOIN users u ON u.id = lr.user_id
		JOIN leave_types lt ON lt.id = lr.leave_type_id
		WHERE lr.created_at >= NOW() - ($1 || ' months')::interval
		  AND lr.status IN ('approved', 'pending')
		ORDER BY lr.start_date DESC
		LIMIT 50
	`, months)
	if err != nil {
		return nil, fmt.Errorf("query team leave history: %w", err)
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var name, leaveType, status string
		var startDate, endDate time.Time
		var totalDays int
		if err := rows.Scan(&name, &leaveType, &startDate, &endDate, &totalDays, &status); err != nil {
			continue
		}
		history = append(history, map[string]interface{}{
			"name":        name,
			"leave_type":  leaveType,
			"start_date":  startDate.Format("2006-01-02"),
			"end_date":    endDate.Format("2006-01-02"),
			"total_days":  totalDays,
			"status":      status,
		})
	}
	return history, nil
}
