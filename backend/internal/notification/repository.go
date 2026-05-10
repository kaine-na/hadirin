package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository menangani akses data untuk modul notification.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat instance Repository baru.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create menyimpan notifikasi baru ke database.
func (r *Repository) Create(ctx context.Context, input *CreateNotificationInput) (*Notification, error) {
	var metaJSON []byte
	if input.Metadata != nil {
		var err error
		metaJSON, err = json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO notifications (user_id, type, title, message, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, type, title, message, is_read, metadata, created_at
	`

	n := &Notification{}
	err := r.db.QueryRow(ctx, query,
		input.UserID,
		input.Type,
		input.Title,
		input.Message,
		metaJSON,
	).Scan(
		&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message,
		&n.IsRead, &n.Metadata, &n.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	return n, nil
}

// List mengambil daftar notifikasi untuk user tertentu.
func (r *Repository) List(ctx context.Context, filter ListFilter) ([]*Notification, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 50 {
		filter.PageSize = 20
	}

	offset := (filter.Page - 1) * filter.PageSize

	whereClause := "WHERE user_id = $1"
	args := []interface{}{filter.UserID}
	argIdx := 2

	if filter.UnreadOnly {
		whereClause += fmt.Sprintf(" AND is_read = false")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications %s", whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	args = append(args, filter.PageSize, offset)
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, type, title, message, is_read, metadata, created_at
		FROM notifications
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		n := &Notification{}
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message,
			&n.IsRead, &n.Metadata, &n.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, total, nil
}

// MarkAsRead menandai satu notifikasi sebagai sudah dibaca.
func (r *Repository) MarkAsRead(ctx context.Context, id, userID string) error {
	query := `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("mark as read: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("notifikasi tidak ditemukan")
	}
	return nil
}

// MarkAllAsRead menandai semua notifikasi user sebagai sudah dibaca.
func (r *Repository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("mark all as read: %w", err)
	}
	return nil
}

// GetUnreadCount mengambil jumlah notifikasi yang belum dibaca untuk user.
func (r *Repository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	var count int
	if err := r.db.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}
	return count, nil
}

// GetUsersNotClockedIn mengambil user_id yang belum clock-in hari ini.
// Digunakan oleh background worker untuk kirim reminder.
func (r *Repository) GetUsersNotClockedIn(ctx context.Context) ([]string, error) {
	query := `
		SELECT u.id
		FROM users u
		WHERE u.is_active = true
		  AND NOT EXISTS (
		    SELECT 1 FROM attendances a
		    WHERE a.user_id = u.id
		      AND DATE(a.clock_in AT TIME ZONE 'Asia/Jakarta') = CURRENT_DATE AT TIME ZONE 'Asia/Jakarta'
		  )
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get users not clocked in: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, nil
}

// GetPendingLeavesOlderThan mengambil leave requests pending yang sudah lebih dari N hari.
// Mengembalikan pasangan (user_id_pemohon, leave_id, manager_id).
func (r *Repository) GetPendingLeavesOlderThan(ctx context.Context, days int) ([]map[string]string, error) {
	query := `
		SELECT lr.id, lr.user_id, u.manager_id
		FROM leave_requests lr
		JOIN users u ON u.id = lr.user_id
		WHERE lr.status = 'pending'
		  AND lr.created_at < NOW() - INTERVAL '1 day' * $1
		  AND u.manager_id IS NOT NULL
	`
	rows, err := r.db.Query(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("get pending leaves: %w", err)
	}
	defer rows.Close()

	var results []map[string]string
	for rows.Next() {
		var leaveID, userID, managerID string
		if err := rows.Scan(&leaveID, &userID, &managerID); err != nil {
			return nil, err
		}
		results = append(results, map[string]string{
			"leave_id":   leaveID,
			"user_id":    userID,
			"manager_id": managerID,
		})
	}
	return results, nil
}

// GetPendingDocsOlderThan mengambil dokumen pending yang sudah lebih dari N hari.
// Mengembalikan pasangan (doc_id, uploader_user_id, hr_user_ids).
func (r *Repository) GetPendingDocsOlderThan(ctx context.Context, days int) ([]map[string]string, error) {
	query := `
		SELECT d.id, d.user_id, hr.id AS hr_id
		FROM documents d
		CROSS JOIN (
		  SELECT id FROM users WHERE role IN ('super_admin', 'hr_admin') AND is_active = true LIMIT 5
		) hr
		WHERE d.status = 'pending'
		  AND d.created_at < NOW() - INTERVAL '1 day' * $1
	`
	rows, err := r.db.Query(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("get pending docs: %w", err)
	}
	defer rows.Close()

	var results []map[string]string
	for rows.Next() {
		var docID, userID, hrID string
		if err := rows.Scan(&docID, &userID, &hrID); err != nil {
			return nil, err
		}
		results = append(results, map[string]string{
			"doc_id":  docID,
			"user_id": userID,
			"hr_id":   hrID,
		})
	}
	return results, nil
}

// HasReminderSentToday mengecek apakah reminder tipe tertentu sudah dikirim ke user hari ini.
// Mencegah duplikasi reminder dari background worker.
func (r *Repository) HasReminderSentToday(ctx context.Context, userID string, notifType NotificationType) (bool, error) {
	query := `
		SELECT EXISTS (
		  SELECT 1 FROM notifications
		  WHERE user_id = $1
		    AND type = $2
		    AND DATE(created_at AT TIME ZONE 'Asia/Jakarta') = CURRENT_DATE AT TIME ZONE 'Asia/Jakarta'
		)
	`
	var exists bool
	if err := r.db.QueryRow(ctx, query, userID, notifType).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
