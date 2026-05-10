package employee

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository menangani semua operasi database untuk karyawan.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat instance Repository baru.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FindByID mengambil satu karyawan berdasarkan ID.
func (r *Repository) FindByID(ctx context.Context, id string) (*Employee, error) {
	emp := &Employee{}
	var joinedAt *time.Time

	err := r.db.QueryRow(ctx, `
		SELECT id, company_id::text, name, email, role,
		       COALESCE(department, ''), COALESCE(position, ''),
		       COALESCE(nik, ''), COALESCE(photo_url, ''),
		       joined_at, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&emp.ID, &emp.CompanyID, &emp.Name, &emp.Email, &emp.Role,
		&emp.Department, &emp.Position, &emp.NIK, &emp.PhotoURL,
		&joinedAt, &emp.IsActive, &emp.CreatedAt, &emp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}
	emp.JoinedAt = joinedAt
	return emp, nil
}

// FindByEmail mengambil user berdasarkan email (untuk cek duplikat).
func (r *Repository) FindByEmail(ctx context.Context, email string) (*Employee, error) {
	emp := &Employee{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, email FROM users WHERE email = $1
	`, email).Scan(&emp.ID, &emp.Name, &emp.Email)
	if err != nil {
		return nil, err
	}
	return emp, nil
}

// List mengambil daftar karyawan dengan filter dan pagination.
func (r *Repository) List(ctx context.Context, filter ListFilter) ([]*Employee, int, error) {
	// Build WHERE clause dinamis
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.Department != "" {
		conditions = append(conditions, fmt.Sprintf("department = $%d", argIdx))
		args = append(args, filter.Department)
		argIdx++
	}
	if filter.Role != "" {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, filter.Role)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx+1))
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
		argIdx += 2
	}

	where := strings.Join(conditions, " AND ")

	// Count total
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE "+where, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count employees: %w", err)
	}

	// Pagination
	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, company_id::text, name, email, role,
		       COALESCE(department, ''), COALESCE(position, ''),
		       COALESCE(nik, ''), COALESCE(photo_url, ''),
		       joined_at, is_active, created_at, updated_at
		FROM users WHERE %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()

	var employees []*Employee
	for rows.Next() {
		emp := &Employee{}
		var joinedAt *time.Time
		if err := rows.Scan(
			&emp.ID, &emp.CompanyID, &emp.Name, &emp.Email, &emp.Role,
			&emp.Department, &emp.Position, &emp.NIK, &emp.PhotoURL,
			&joinedAt, &emp.IsActive, &emp.CreatedAt, &emp.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan employee: %w", err)
		}
		emp.JoinedAt = joinedAt
		employees = append(employees, emp)
	}

	return employees, total, nil
}

// Create membuat karyawan baru di database.
func (r *Repository) Create(ctx context.Context, emp *Employee, passwordHash string) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, role, department, position, nik, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		RETURNING id, created_at, updated_at
	`, emp.Name, emp.Email, passwordHash, emp.Role,
		nullableString(emp.Department), nullableString(emp.Position), nullableString(emp.NIK),
	).Scan(&emp.ID, &emp.CreatedAt, &emp.UpdatedAt)
}

// Update mengupdate data karyawan.
func (r *Repository) Update(ctx context.Context, id string, req *UpdateEmployeeRequest) (*Employee, error) {
	emp := &Employee{}
	var joinedAt *time.Time

	err := r.db.QueryRow(ctx, `
		UPDATE users SET
			name = COALESCE(NULLIF($2, ''), name),
			role = COALESCE(NULLIF($3, ''), role),
			department = COALESCE(NULLIF($4, ''), department),
			position = COALESCE(NULLIF($5, ''), position),
			nik = COALESCE(NULLIF($6, ''), nik),
			is_active = COALESCE($7, is_active),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, company_id::text, name, email, role,
		          COALESCE(department, ''), COALESCE(position, ''),
		          COALESCE(nik, ''), COALESCE(photo_url, ''),
		          joined_at, is_active, created_at, updated_at
	`, id, req.Name, req.Role, req.Department, req.Position, req.NIK, req.IsActive,
	).Scan(
		&emp.ID, &emp.CompanyID, &emp.Name, &emp.Email, &emp.Role,
		&emp.Department, &emp.Position, &emp.NIK, &emp.PhotoURL,
		&joinedAt, &emp.IsActive, &emp.CreatedAt, &emp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update employee: %w", err)
	}
	emp.JoinedAt = joinedAt
	return emp, nil
}

// SoftDelete menonaktifkan karyawan (is_active = false).
func (r *Repository) SoftDelete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `
		UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete employee: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("employee not found")
	}
	return nil
}

// UpdatePhoto mengupdate URL foto profil karyawan.
func (r *Repository) UpdatePhoto(ctx context.Context, id, photoURL string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET photo_url = $2, updated_at = NOW() WHERE id = $1
	`, id, photoURL)
	return err
}

// nullableString mengembalikan nil jika string kosong, untuk INSERT/UPDATE.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
