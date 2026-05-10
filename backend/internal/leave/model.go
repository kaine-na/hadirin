package leave

import "time"

// LeaveStatus adalah enum status pengajuan cuti.
type LeaveStatus string

const (
	StatusPending   LeaveStatus = "pending"
	StatusApproved  LeaveStatus = "approved"
	StatusRejected  LeaveStatus = "rejected"
	StatusCancelled LeaveStatus = "cancelled"
)

// LeaveType adalah jenis cuti yang tersedia.
type LeaveType struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	MaxDays     int       `json:"max_days"`
	IsPaid      bool      `json:"is_paid"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LeaveBalance adalah saldo cuti per karyawan per tahun.
type LeaveBalance struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	LeaveTypeID   string    `json:"leave_type_id"`
	LeaveTypeName string    `json:"leave_type_name,omitempty"`
	Year          int       `json:"year"`
	TotalDays     int       `json:"total_days"`
	UsedDays      int       `json:"used_days"`
	RemainingDays int       `json:"remaining_days"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// LeaveRequest adalah pengajuan cuti karyawan.
type LeaveRequest struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	UserName        string      `json:"user_name,omitempty"`
	LeaveTypeID     string      `json:"leave_type_id"`
	LeaveTypeName   string      `json:"leave_type_name,omitempty"`
	StartDate       time.Time   `json:"start_date"`
	EndDate         time.Time   `json:"end_date"`
	TotalDays       int         `json:"total_days"`
	Reason          string      `json:"reason"`
	Status          LeaveStatus `json:"status"`
	ApprovedBy      string      `json:"approved_by,omitempty"`
	ApprovedByName  string      `json:"approved_by_name,omitempty"`
	ApprovedAt      *time.Time  `json:"approved_at,omitempty"`
	RejectionReason string      `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// CreateLeaveRequest adalah body request untuk mengajukan cuti.
type CreateLeaveRequest struct {
	LeaveTypeID string `json:"leave_type_id"`
	StartDate   string `json:"start_date"` // YYYY-MM-DD
	EndDate     string `json:"end_date"`   // YYYY-MM-DD
	Reason      string `json:"reason"`
}

// ApproveRequest adalah body request untuk approve cuti.
type ApproveRequest struct {
	// Tidak ada field tambahan, hanya aksi approve
}

// RejectRequest adalah body request untuk reject cuti.
type RejectRequest struct {
	RejectionReason string `json:"rejection_reason"`
}

// LeaveFilter adalah parameter filter untuk list cuti.
type LeaveFilter struct {
	UserID    string
	Status    string
	StartDate string
	EndDate   string
	Page      int
	PageSize  int
}

// AIRecommendation adalah hasil rekomendasi AI untuk pengajuan cuti.
type AIRecommendation struct {
	Recommendation string `json:"recommendation"` // "Direkomendasikan disetujui" / "Perlu pertimbangan"
	Reason         string `json:"reason"`
}
