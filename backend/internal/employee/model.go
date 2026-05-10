package employee

import "time"

// Employee adalah representasi data karyawan dari database.
type Employee struct {
	ID         string     `json:"id"`
	CompanyID  string     `json:"company_id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Department string     `json:"department,omitempty"`
	Position   string     `json:"position,omitempty"`
	NIK        string     `json:"nik,omitempty"`
	PhotoURL   string     `json:"photo_url,omitempty"`
	JoinedAt   *time.Time `json:"joined_at,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// CreateEmployeeRequest adalah body request untuk membuat karyawan baru.
type CreateEmployeeRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	Department string `json:"department"`
	Position   string `json:"position"`
	NIK        string `json:"nik"`
}

// UpdateEmployeeRequest adalah body request untuk update data karyawan.
type UpdateEmployeeRequest struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	Department string `json:"department"`
	Position   string `json:"position"`
	NIK        string `json:"nik"`
	IsActive   *bool  `json:"is_active"`
}

// ListFilter adalah parameter filter untuk list karyawan.
type ListFilter struct {
	Department string
	Role       string
	Search     string
	Page       int
	PageSize   int
}
