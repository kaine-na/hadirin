package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ChecklistItem menyimpan satu item checklist kepatuhan.
type ChecklistItem struct {
	ID          string     `json:"id"`
	Period      string     `json:"period"`
	ItemCode    string     `json:"item_code"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Deadline    time.Time  `json:"deadline"`
	Status      string     `json:"status"` // pending, done, overdue
	DoneAt      *time.Time `json:"done_at,omitempty"`
	DoneBy      *string    `json:"done_by,omitempty"`
	NotifiedH3  bool       `json:"notified_h3"`
	DaysUntil   int        `json:"days_until"` // hari tersisa (negatif = overdue)
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// checklistTemplate mendefinisikan template item checklist bulanan.
type checklistTemplate struct {
	Code        string
	Title       string
	Description string
	DayOfMonth  int // tanggal deadline dalam bulan
}

// defaultChecklistTemplates adalah daftar kewajiban compliance bulanan.
var defaultChecklistTemplates = []checklistTemplate{
	{
		Code:        "LAPOR_BPJS_KESEHATAN",
		Title:       "Lapor & Bayar BPJS Kesehatan",
		Description: "Pembayaran iuran BPJS Kesehatan paling lambat tanggal 10 bulan berikutnya",
		DayOfMonth:  10,
	},
	{
		Code:        "LAPOR_BPJS_TK",
		Title:       "Lapor & Bayar BPJS Ketenagakerjaan",
		Description: "Pembayaran iuran BPJS Ketenagakerjaan (JHT, JP, JKK, JKM) paling lambat tanggal 15",
		DayOfMonth:  15,
	},
	{
		Code:        "SETOR_PPH21",
		Title:       "Setor PPh 21",
		Description: "Penyetoran PPh 21 karyawan ke kas negara paling lambat tanggal 10 bulan berikutnya",
		DayOfMonth:  10,
	},
	{
		Code:        "LAPOR_SPT_MASA_PPH21",
		Title:       "Lapor SPT Masa PPh 21",
		Description: "Pelaporan SPT Masa PPh 21 ke DJP paling lambat tanggal 20 bulan berikutnya",
		DayOfMonth:  20,
	},
	{
		Code:        "REKAP_GAJI",
		Title:       "Rekap & Verifikasi Gaji",
		Description: "Rekap dan verifikasi data gaji karyawan sebelum proses payroll",
		DayOfMonth:  25,
	},
}

// ChecklistRepository menangani operasi database untuk checklist.
type ChecklistRepository struct {
	db *pgxpool.Pool
}

// NewChecklistRepository membuat instance ChecklistRepository baru.
func NewChecklistRepository(db *pgxpool.Pool) *ChecklistRepository {
	return &ChecklistRepository{db: db}
}

// GenerateMonthlyChecklist membuat checklist untuk bulan tertentu jika belum ada.
// period format: "YYYY-MM"
func (r *ChecklistRepository) GenerateMonthlyChecklist(ctx context.Context, period string) ([]*ChecklistItem, error) {
	// Parse period
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return nil, fmt.Errorf("format period tidak valid, gunakan YYYY-MM: %w", err)
	}

	// Deadline ada di bulan berikutnya untuk beberapa item
	nextMonth := t.AddDate(0, 1, 0)

	items := make([]*ChecklistItem, 0, len(defaultChecklistTemplates))
	for _, tmpl := range defaultChecklistTemplates {
		// Deadline: tanggal tertentu di bulan berikutnya
		deadline := time.Date(nextMonth.Year(), nextMonth.Month(), tmpl.DayOfMonth, 23, 59, 59, 0, time.UTC)

		// Upsert: insert jika belum ada, skip jika sudah ada
		var item ChecklistItem
		err := r.db.QueryRow(ctx, `
			INSERT INTO compliance_checklist (period, item_code, title, description, deadline, status)
			VALUES ($1, $2, $3, $4, $5, 'pending')
			ON CONFLICT (period, item_code) DO UPDATE
				SET updated_at = NOW()
			RETURNING id, period, item_code, title, description, deadline, status,
				done_at, done_by, notified_h3, created_at, updated_at
		`, period, tmpl.Code, tmpl.Title, tmpl.Description, deadline).Scan(
			&item.ID, &item.Period, &item.ItemCode, &item.Title, &item.Description,
			&item.Deadline, &item.Status, &item.DoneAt, &item.DoneBy,
			&item.NotifiedH3, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal generate checklist item %s: %w", tmpl.Code, err)
		}

		item.DaysUntil = int(time.Until(item.Deadline).Hours() / 24)
		items = append(items, &item)
	}

	return items, nil
}

// GetChecklist mengambil checklist untuk periode tertentu.
func (r *ChecklistRepository) GetChecklist(ctx context.Context, period string) ([]*ChecklistItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, period, item_code, title, description, deadline, status,
			done_at, done_by, notified_h3, created_at, updated_at
		FROM compliance_checklist
		WHERE period = $1
		ORDER BY deadline ASC
	`, period)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil checklist: %w", err)
	}
	defer rows.Close()

	var items []*ChecklistItem
	for rows.Next() {
		var item ChecklistItem
		if err := rows.Scan(
			&item.ID, &item.Period, &item.ItemCode, &item.Title, &item.Description,
			&item.Deadline, &item.Status, &item.DoneAt, &item.DoneBy,
			&item.NotifiedH3, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.DaysUntil = int(time.Until(item.Deadline).Hours() / 24)
		items = append(items, &item)
	}

	return items, rows.Err()
}

// MarkDone menandai satu item checklist sebagai selesai.
func (r *ChecklistRepository) MarkDone(ctx context.Context, id, doneByUserID string) (*ChecklistItem, error) {
	var item ChecklistItem
	err := r.db.QueryRow(ctx, `
		UPDATE compliance_checklist
		SET status = 'done', done_at = NOW(), done_by = $2, updated_at = NOW()
		WHERE id = $1 AND status != 'done'
		RETURNING id, period, item_code, title, description, deadline, status,
			done_at, done_by, notified_h3, created_at, updated_at
	`, id, doneByUserID).Scan(
		&item.ID, &item.Period, &item.ItemCode, &item.Title, &item.Description,
		&item.Deadline, &item.Status, &item.DoneAt, &item.DoneBy,
		&item.NotifiedH3, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal menandai checklist selesai: %w", err)
	}
	item.DaysUntil = int(time.Until(item.Deadline).Hours() / 24)
	return &item, nil
}

// UpdateOverdueItems memperbarui status item yang sudah melewati deadline.
func (r *ChecklistRepository) UpdateOverdueItems(ctx context.Context) (int64, error) {
	result, err := r.db.Exec(ctx, `
		UPDATE compliance_checklist
		SET status = 'overdue', updated_at = NOW()
		WHERE status = 'pending' AND deadline < NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("gagal update overdue items: %w", err)
	}
	return result.RowsAffected(), nil
}

// GetItemsDueInDays mengambil item yang deadline-nya dalam N hari ke depan.
// Digunakan untuk notifikasi H-3.
func (r *ChecklistRepository) GetItemsDueInDays(ctx context.Context, days int) ([]*ChecklistItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, period, item_code, title, description, deadline, status,
			done_at, done_by, notified_h3, created_at, updated_at
		FROM compliance_checklist
		WHERE status = 'pending'
			AND deadline BETWEEN NOW() AND NOW() + ($1 || ' days')::INTERVAL
			AND notified_h3 = false
		ORDER BY deadline ASC
	`, days)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil items due in %d days: %w", days, err)
	}
	defer rows.Close()

	var items []*ChecklistItem
	for rows.Next() {
		var item ChecklistItem
		if err := rows.Scan(
			&item.ID, &item.Period, &item.ItemCode, &item.Title, &item.Description,
			&item.Deadline, &item.Status, &item.DoneAt, &item.DoneBy,
			&item.NotifiedH3, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.DaysUntil = int(time.Until(item.Deadline).Hours() / 24)
		items = append(items, &item)
	}

	return items, rows.Err()
}

// MarkNotifiedH3 menandai item sudah dikirim notifikasi H-3.
func (r *ChecklistRepository) MarkNotifiedH3(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE compliance_checklist
		SET notified_h3 = true, updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}
